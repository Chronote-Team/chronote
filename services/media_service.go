package services

import (
	"bytes"
	"chronote/config"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"chronote/global"
	"chronote/models"
	"chronote/utils"
)

type MediaService struct{}

var allowedMediaGroups = map[string]bool{
	"header":  true,
	"gallery": true,
	"bgm":     true,
}

func (s *MediaService) ProcessAndUpload(file *multipart.FileHeader, postcardID uint, mediaType, mediaGroup string) (*models.PostcardMedia, error) {
	return s.processAndUploadWithDB(global.Db, file, postcardID, mediaType, mediaGroup, 0)
}

func (s *MediaService) BatchProcessAndUpload(postcardID uint, files []*multipart.FileHeader, mediaType, mediaGroup string) ([]models.PostcardMedia, error) {
	if len(files) == 0 {
		return nil, errors.New("请上传媒体文件")
	}

	uploadedMedias := make([]models.PostcardMedia, 0, len(files))
	cleanupDone := false

	err := global.Db.Transaction(func(tx *gorm.DB) error {
		startPosition, err := s.nextPositionWithDB(tx, postcardID)
		if err != nil {
			return errors.New("获取媒体排序失败")
		}

		for index, file := range files {
			media, err := s.uploadMediaFile(file, postcardID, mediaType, mediaGroup, startPosition+index)
			if err != nil {
				cleanupMediaObjects(uploadedMedias)
				cleanupDone = true
				return err
			}
			uploadedMedias = append(uploadedMedias, *media)
		}

		if err := tx.Create(&uploadedMedias).Error; err != nil {
			cleanupMediaObjects(uploadedMedias)
			cleanupDone = true
			return errors.New("保存媒体信息失败")
		}
		return nil
	})
	if err != nil {
		if !cleanupDone {
			cleanupMediaObjects(uploadedMedias)
		}
		return nil, err
	}
	return uploadedMedias, nil
}

func (s *MediaService) processAndUploadWithDB(db *gorm.DB, file *multipart.FileHeader, postcardID uint, mediaType, mediaGroup string, position int) (*models.PostcardMedia, error) {
	var err error
	if position <= 0 {
		position, err = s.nextPositionWithDB(db, postcardID)
		if err != nil {
			return nil, errors.New("获取媒体排序失败")
		}
	}

	media, err := s.uploadMediaFile(file, postcardID, mediaType, mediaGroup, position)
	if err != nil {
		return nil, err
	}

	if err := db.Create(media).Error; err != nil {
		_ = deleteMediaObjects(*media)
		return nil, errors.New("保存媒体信息失败")
	}
	return media, nil
}

func (s *MediaService) uploadMediaFile(file *multipart.FileHeader, postcardID uint, mediaType, mediaGroup string, position int) (*models.PostcardMedia, error) {
	if file == nil {
		return nil, errors.New("媒体文件不能为空")
	}

	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	if mediaType != "" && mediaType != "image" && mediaType != "video" && mediaType != "audio" {
		return nil, errors.New("媒体类型无效")
	}

	actualMediaType, detectedContentType, err := detectMediaTypeByMagicBytes(file)
	if err != nil {
		return nil, err
	}

	if mediaType == "" {
		mediaType = actualMediaType
	} else if mediaType != actualMediaType {
		return nil, errors.New("文件类型与声明不符")
	}

	maxSize := getMediaMaxSize(mediaType)
	if err := validateMediaFileSize(file.Size, maxSize); err != nil {
		return nil, err
	}

	mediaGroup = normalizeMediaGroup(mediaGroup)
	if mediaGroup == "" {
		mediaGroup = "gallery"
	}
	if !allowedMediaGroups[mediaGroup] {
		return nil, errors.New("媒体分组无效")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	objectKey := fmt.Sprintf("postcards/%d/%d%s", postcardID, time.Now().UnixNano(), ext)
	contentType := detectedContentType

	media := models.PostcardMedia{
		PostcardID: postcardID,
		MediaType:  mediaType,
		OSSKey:     objectKey,
		FileSize:   file.Size,
		MediaGroup: mediaGroup,
		Position:   position,
	}

	switch mediaType {
	case "image":
		data, err := readFileData(file, maxSize)
		if err != nil {
			return nil, err
		}
		url, err := utils.UploadPostcardObject(objectKey, file.Filename, bytes.NewReader(data), contentType)
		if err != nil {
			return nil, errors.New("上传媒体失败")
		}
		media.URL = url
		if width, height, err := utils.GetImageDimensions(data); err == nil {
			media.OriginalWidth = width
			media.OriginalHeight = height
		}
		if thumbnailData, err := utils.GenerateImageThumbnail(data, 480); err == nil {
			thumbnailKey := fmt.Sprintf("postcards/%d/thumbnails/%d%s", postcardID, time.Now().UnixNano(), ext)
			thumbnailURL, err := utils.UploadPostcardObject(thumbnailKey, file.Filename, bytes.NewReader(thumbnailData), contentType)
			if err == nil {
				media.ThumbnailURL = thumbnailURL
				media.ThumbnailOSSKey = thumbnailKey
			}
		}
	default:
		src, err := file.Open()
		if err != nil {
			return nil, errors.New("读取媒体失败")
		}
		defer src.Close()
		url, err := utils.UploadPostcardObject(objectKey, file.Filename, src, contentType)
		if err != nil {
			return nil, errors.New("上传媒体失败")
		}
		media.URL = url
	}

	return &media, nil
}

func (s *MediaService) Delete(postcardID, mediaID uint) error {
	var media models.PostcardMedia
	if err := global.Db.Where("postcard_id = ? AND id = ?", postcardID, mediaID).First(&media).Error; err != nil {
		return errors.New("媒体不存在")
	}
	if err := deleteMediaObjects(media); err != nil {
		return err
	}
	if err := global.Db.Delete(&media).Error; err != nil {
		return errors.New("删除媒体失败")
	}
	return nil
}

func (s *MediaService) Reorder(postcardID uint, mediaIDs []uint) error {
	if len(mediaIDs) == 0 {
		return errors.New("媒体排序不能为空")
	}
	var total int64
	if err := global.Db.Model(&models.PostcardMedia{}).
		Where("postcard_id = ?", postcardID).
		Count(&total).Error; err != nil {
		return errors.New("获取媒体失败")
	}
	if total != int64(len(mediaIDs)) {
		return errors.New("必须传入全部媒体ID进行排序")
	}
	var count int64
	if err := global.Db.Model(&models.PostcardMedia{}).
		Where("postcard_id = ? AND id IN ?", postcardID, mediaIDs).
		Count(&count).Error; err != nil {
		return errors.New("获取媒体失败")
	}
	if count != int64(len(mediaIDs)) {
		return errors.New("媒体列表不匹配")
	}
	return global.Db.Transaction(func(tx *gorm.DB) error {
		for index, mediaID := range mediaIDs {
			if err := tx.Model(&models.PostcardMedia{}).
				Where("postcard_id = ? AND id = ?", postcardID, mediaID).
				Update("position", index+1).Error; err != nil {
				return errors.New("更新媒体排序失败")
			}
		}
		return nil
	})
}

func (s *MediaService) List(postcardID uint) ([]models.PostcardMedia, error) {
	var medias []models.PostcardMedia
	if err := global.Db.Where("postcard_id = ?", postcardID).
		Order("position asc").
		Find(&medias).Error; err != nil {
		return nil, errors.New("获取媒体列表失败")
	}
	return medias, nil
}

func (s *MediaService) nextPosition(postcardID uint) (int, error) {
	return s.nextPositionWithDB(global.Db, postcardID)
}

func (s *MediaService) nextPositionWithDB(db *gorm.DB, postcardID uint) (int, error) {
	var maxPosition int
	if err := db.Model(&models.PostcardMedia{}).
		Where("postcard_id = ?", postcardID).
		Select("COALESCE(MAX(position), 0)").Scan(&maxPosition).Error; err != nil {
		return 0, err
	}
	return maxPosition + 1, nil
}

func normalizeMediaGroup(mediaGroup string) string {
	return strings.ToLower(strings.TrimSpace(mediaGroup))
}

func readFileData(file *multipart.FileHeader, maxSize int64) ([]byte, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	limitedReader := io.LimitReader(src, maxSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, errors.New("读取媒体失败")
	}
	if int64(len(data)) > maxSize {
		return nil, errors.New("媒体文件大小超出限制")
	}
	return data, nil
}

func getMediaMaxSize(mediaType string) int64 {
	const (
		defaultMaxImageSize int64 = 10 * 1024 * 1024
		defaultMaxVideoSize int64 = 200 * 1024 * 1024
		defaultMaxAudioSize int64 = 50 * 1024 * 1024
	)

	switch mediaType {
	case "image":
		if config.AppConfig != nil && config.AppConfig.Media.MaxImageSize > 0 {
			return config.AppConfig.Media.MaxImageSize
		}
		return defaultMaxImageSize
	case "video":
		if config.AppConfig != nil && config.AppConfig.Media.MaxVideoSize > 0 {
			return config.AppConfig.Media.MaxVideoSize
		}
		return defaultMaxVideoSize
	case "audio":
		if config.AppConfig != nil && config.AppConfig.Media.MaxAudioSize > 0 {
			return config.AppConfig.Media.MaxAudioSize
		}
		return defaultMaxAudioSize
	default:
		return defaultMaxImageSize
	}
}

func validateMediaFileSize(fileSize, maxSize int64) error {
	if maxSize <= 0 {
		return nil
	}
	if fileSize > maxSize {
		return errors.New("媒体文件大小超出限制")
	}
	return nil
}

func detectMediaTypeByMagicBytes(file *multipart.FileHeader) (string, string, error) {
	src, err := file.Open()
	if err != nil {
		return "", "", errors.New("读取媒体失败")
	}
	defer src.Close()

	buffer := make([]byte, 512)
	n, err := io.ReadFull(src, buffer)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", "", errors.New("读取媒体失败")
	}
	if n == 0 {
		return "", "", errors.New("媒体文件不能为空")
	}

	contentType := strings.ToLower(http.DetectContentType(buffer[:n]))
	switch {
	case strings.HasPrefix(contentType, "image/"):
		return "image", contentType, nil
	case strings.HasPrefix(contentType, "video/"):
		return "video", contentType, nil
	case strings.HasPrefix(contentType, "audio/"):
		return "audio", contentType, nil
	case contentType == "application/ogg":
		return "audio", contentType, nil
	default:
		return "", "", errors.New("不支持的媒体类型")
	}
}

func deleteMediaObjects(media models.PostcardMedia) error {
	if err := utils.DeleteObject(media.OSSKey); err != nil {
		return errors.New("删除媒体文件失败")
	}
	if media.ThumbnailOSSKey != "" && media.ThumbnailOSSKey != media.OSSKey {
		if err := utils.DeleteObject(media.ThumbnailOSSKey); err != nil {
			return errors.New("删除缩略图失败")
		}
	}
	return nil
}
