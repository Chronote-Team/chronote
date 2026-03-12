package services

import (
	"encoding/json"
	"errors"
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
	"friends": true,
}

func (s *PostcardService) Create(userID uint, req *models.CreatePostcardRequest, files []*multipart.FileHeader, mediaType, mediaGroup string) (*models.Postcard, error) {
	visibility := normalizeVisibility(req.Visibility)
	if visibility == "" {
		visibility = "private"
	}
	if err := validateVisibility(visibility); err != nil {
		return nil, err
	}
	if !json.Valid(req.Content) {
		return nil, errors.New("content 无效")
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, errors.New("title 不能为空")
	}

	postcard := models.Postcard{
		Title:      title,
		Content:    datatypes.JSON(req.Content),
		Visibility: visibility,
		AuthorID:   userID,
	}
	mediaService := MediaService{}
	uploadedMedias := make([]models.PostcardMedia, 0, len(files))

	err := global.Db.Transaction(func(tx *gorm.DB) error {
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
		if !json.Valid(*req.Content) {
			return errors.New("content 无效")
		}
		updates["content"] = datatypes.JSON(*req.Content)
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
	for _, media := range postcard.Medias {
		if err := deleteMediaObjects(media); err != nil {
			return err
		}
	}
	if err := global.Db.Where("postcard_id = ?", postcardID).Delete(&models.PostcardMedia{}).Error; err != nil {
		return errors.New("删除媒体失败")
	}
	if err := global.Db.Delete(&models.Postcard{}, postcardID).Error; err != nil {
		return errors.New("删除明信片失败")
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
	case "friends":
		return false
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

func cleanupMediaObjects(medias []models.PostcardMedia) {
	for _, media := range medias {
		_ = deleteMediaObjects(media)
	}
}
