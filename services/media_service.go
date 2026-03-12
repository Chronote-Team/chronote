package services

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
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

func (s *MediaService) processAndUploadWithDB(db *gorm.DB, file *multipart.FileHeader, postcardID uint, mediaType, mediaGroup string, position int) (*models.PostcardMedia, error) {
	if file == nil {
		return nil, errors.New("媒体文件不能为空")
	}
	if mediaType == "" {
		detectedType, err := utils.DetectMediaType(file.Filename, file.Header.Get("Content-Type"))
		if err != nil {
			return nil, err
		}
		mediaType = detectedType
	} else {
		mediaType = strings.ToLower(strings.TrimSpace(mediaType))
		if mediaType != "image" && mediaType != "video" && mediaType != "audio" {
			return nil, errors.New("媒体类型无效")
		}
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
	contentType := file.Header.Get("Content-Type")

	media := models.PostcardMedia{
		PostcardID: postcardID,
		MediaType:  mediaType,
		OSSKey:     objectKey,
		FileSize:   file.Size,
		MediaGroup: mediaGroup,
	}
	var err error
	if position <= 0 {
		position, err = s.nextPositionWithDB(db, postcardID)
		if err != nil {
			return nil, errors.New("获取媒体排序失败")
		}
	}
	media.Position = position

	switch mediaType {
	case "image":
		data, err := readFileData(file)
		if err != nil {
			return nil, errors.New("读取媒体失败")
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

	if err := db.Create(&media).Error; err != nil {
		_ = deleteMediaObjects(media)
		return nil, errors.New("保存媒体信息失败")
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

func readFileData(file *multipart.FileHeader) ([]byte, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()
	return io.ReadAll(src)
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
