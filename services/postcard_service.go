package services

import (
	"encoding/json"
	"errors"
	"log"
	"mime/multipart"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"chronote/global"
	"chronote/models"
)

type PostcardService struct{}

var allowedVisibility = map[string]bool{
	"public":  true,
	"private": true,
}

const maxPostcardContentBytes = 64 * 1024

func (s *PostcardService) Create(userID uint, req *models.CreatePostcardRequest, files []*multipart.FileHeader, mediaType, mediaGroup string) (*models.Postcard, error) {
	if len(files) > maxMediaFilesPerRequest {
		return nil, errors.New("单次最多上传 10 个媒体文件")
	}
	if len(files) > maxMediasPerPostcard {
		return nil, errors.New("单张明信片最多允许 20 个媒体文件")
	}
	normalizedGroup := normalizeMediaGroup(mediaGroup)
	if normalizedGroup == "" {
		normalizedGroup = "gallery"
	}
	if (normalizedGroup == "header" || normalizedGroup == "bgm") && len(files) > 1 {
		return nil, errors.New(normalizedGroup + " 分组一次只能上传 1 个文件")
	}

	visibility := normalizeVisibility(req.Visibility)
	if visibility == "" {
		visibility = "private"
	}
	if err := validateVisibility(visibility); err != nil {
		return nil, err
	}
	title, err := validatePostcardTitle(req.Title)
	if err != nil {
		return nil, err
	}
	content, err := validatePostcardContent(req.Content)
	if err != nil {
		return nil, err
	}

	postcard := models.Postcard{
		Title:      title,
		Content:    content,
		Visibility: visibility,
		AuthorID:   userID,
	}
	mediaService := MediaService{}
	uploadedMedias := make([]models.PostcardMedia, 0, len(files))

	err = global.Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&postcard).Error; err != nil {
			return errors.New("创建明信片失败")
		}
		for index, file := range files {
			media, err := mediaService.processAndUploadWithDB(tx, file, postcard.ID, mediaType, mediaGroup, index+1)
			if err != nil {
				return err
			}
			uploadedMedias = append(uploadedMedias, *media)
		}
		postcard.Medias = append(postcard.Medias, uploadedMedias...)
		return nil
	})
	if err != nil {
		cleanupMediaObjects(uploadedMedias)
		return nil, err
	}
	return &postcard, nil
}

func (s *PostcardService) List(userID uint, query models.PostcardListQuery) ([]models.Postcard, models.Pagination, error) {
	page, pageSize := normalizePagination(query.Page, query.PageSize)
	sortBy, order := normalizeSort(query.SortBy, query.Order)
	db := global.Db.Model(&models.Postcard{}).
		Preload("Medias", func(db *gorm.DB) *gorm.DB {
			return db.Order("position asc")
		}).
		Preload("Author").
		Where("author_id = ? OR visibility = ?", userID, "public")
	if query.Visibility != "" {
		visibility := normalizeVisibility(query.Visibility)
		if err := validateVisibility(visibility); err != nil {
			return nil, models.Pagination{}, err
		}
		db = db.Where("visibility = ?", visibility)
	}
	var total int64
	if err := db.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, models.Pagination{}, errors.New("获取明信片列表失败")
	}
	var postcards []models.Postcard
	if err := db.Order(sortBy + " " + order).
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&postcards).Error; err != nil {
		return nil, models.Pagination{}, errors.New("获取明信片列表失败")
	}
	return postcards, models.Pagination{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (s *PostcardService) GetDetail(userID, postcardID uint) (*models.Postcard, error) {
	postcard, err := s.getPostcardWithRelations(postcardID)
	if err != nil {
		return nil, err
	}
	if !canAccessPostcard(userID, postcard.AuthorID, postcard.Visibility) {
		return nil, errors.New("无权限访问该明信片")
	}
	return postcard, nil
}

func (s *PostcardService) Update(userID, postcardID uint, req *models.UpdatePostcardRequest) error {
	postcard, err := s.getPostcardByID(postcardID)
	if err != nil {
		return err
	}
	if postcard.AuthorID != userID {
		return errors.New("无权限操作该明信片")
	}
	updates := make(map[string]interface{})
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			return errors.New("title 不能为空")
		}
		updates["title"] = title
	}
	if req.Content != nil {
		content, err := validatePostcardContent(*req.Content)
		if err != nil {
			return err
		}
		updates["content"] = content
	}
	if req.Visibility != nil {
		visibility := normalizeVisibility(*req.Visibility)
		if visibility == "" {
			return errors.New("visibility 无效")
		}
		if err := validateVisibility(visibility); err != nil {
			return err
		}
		updates["visibility"] = visibility
	}
	if len(updates) == 0 {
		return errors.New("没有可更新的字段")
	}
	if err := global.Db.Model(&postcard).Updates(updates).Error; err != nil {
		return errors.New("更新明信片失败")
	}
	return nil
}

func (s *PostcardService) Delete(userID, postcardID uint) error {
	postcard, err := s.getPostcardWithRelations(postcardID)
	if err != nil {
		return err
	}
	if postcard.AuthorID != userID {
		return errors.New("无权限操作该明信片")
	}
	if err := global.Db.Where("postcard_id = ?", postcardID).Delete(&models.PostcardMedia{}).Error; err != nil {
		return errors.New("删除媒体失败")
	}
	if err := global.Db.Delete(&models.Postcard{}, postcardID).Error; err != nil {
		return errors.New("删除明信片失败")
	}
	for _, media := range postcard.Medias {
		if err := deleteMediaObjects(media); err != nil {
			log.Printf("Failed to delete postcard media object after db deletion, postcardID=%d mediaID=%d: %v", postcardID, media.ID, err)
		}
	}
	return nil
}

func (s *PostcardService) EnsureOwner(userID, postcardID uint) (*models.Postcard, error) {
	postcard, err := s.getPostcardByID(postcardID)
	if err != nil {
		return nil, err
	}
	if postcard.AuthorID != userID {
		return nil, errors.New("无权限操作该明信片")
	}
	return postcard, nil
}

func (s *PostcardService) getPostcardByID(postcardID uint) (*models.Postcard, error) {
	var postcard models.Postcard
	if err := global.Db.First(&postcard, postcardID).Error; err != nil {
		return nil, errors.New("明信片不存在")
	}
	return &postcard, nil
}

func (s *PostcardService) getPostcardWithRelations(postcardID uint) (*models.Postcard, error) {
	var postcard models.Postcard
	if err := global.Db.Preload("Medias", func(db *gorm.DB) *gorm.DB {
		return db.Order("position asc")
	}).Preload("Author").First(&postcard, postcardID).Error; err != nil {
		return nil, errors.New("明信片不存在")
	}
	return &postcard, nil
}

func canAccessPostcard(userID, authorID uint, visibility string) bool {
	if userID == authorID {
		return true
	}
	switch visibility {
	case "public":
		return true
	default:
		return false
	}
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}
	return page, pageSize
}

func normalizeSort(sortBy, order string) (string, string) {
	sortBy = strings.ToLower(sortBy)
	switch sortBy {
	case "updated_at":
		sortBy = "updated_at"
	default:
		sortBy = "created_at"
	}
	order = strings.ToLower(order)
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	return sortBy, order
}

func normalizeVisibility(visibility string) string {
	return strings.ToLower(strings.TrimSpace(visibility))
}

func validateVisibility(visibility string) error {
	if visibility == "" {
		return nil
	}
	if !allowedVisibility[visibility] {
		return errors.New("visibility 无效")
	}
	return nil
}

func validatePostcardTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", errors.New("title 不能为空")
	}
	if len(title) > 200 {
		return "", errors.New("title 长度不能超过 200 个字符")
	}
	return title, nil
}

func validatePostcardContent(content json.RawMessage) (datatypes.JSON, error) {
	if len(content) == 0 {
		return nil, errors.New("content 不能为空")
	}
	if len(content) > maxPostcardContentBytes {
		return nil, errors.New("content 长度不能超过 65536 字节")
	}
	if !json.Valid(content) {
		return nil, errors.New("content 无效")
	}

	var decoded interface{}
	if err := json.Unmarshal(content, &decoded); err != nil {
		return nil, errors.New("content 无效")
	}
	switch decoded.(type) {
	case map[string]interface{}, []interface{}:
		return datatypes.JSON(content), nil
	default:
		return nil, errors.New("content 必须是 JSON 对象或数组")
	}
}

func cleanupMediaObjects(medias []models.PostcardMedia) {
	for _, media := range medias {
		_ = deleteMediaObjects(media)
	}
}
