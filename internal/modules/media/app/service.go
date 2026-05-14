package app

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"sort"
	"strings"
	"time"

	mediadomain "chronote-refactor/internal/modules/media/domain"
	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	"chronote-refactor/internal/shared/errs"
)

const (
	maxFilesPerRequest = 10
	maxMediasPerCard   = 20
)

type Service struct {
	repo      Repository
	storage   Storage
	processor ImageProcessor
	ai        AnalysisEnqueuer
}

func NewService(repo Repository, storage Storage, processor ImageProcessor) *Service {
	if repo == nil {
		repo = newMemoryRepository()
	}
	if storage == nil {
		storage = noopStorage{}
	}
	if processor == nil {
		processor = noopImageProcessor{}
	}
	return &Service{repo: repo, storage: storage, processor: processor, ai: postcardaiapp.NoopEnqueuer{}}
}

func (s *Service) SetAnalysisEnqueuer(enqueuer AnalysisEnqueuer) {
	if enqueuer == nil {
		enqueuer = postcardaiapp.NoopEnqueuer{}
	}
	s.ai = enqueuer
}

func (s *Service) Repository() Repository {
	return s.repo
}

func (s *Service) UploadBatch(postcardID uint, files []*multipart.FileHeader, mediaType, group string) ([]mediadomain.Media, error) {
	if len(files) == 0 {
		return nil, errs.Validation("请上传媒体文件")
	}
	if len(files) > maxFilesPerRequest {
		return nil, errs.Validation("单次最多上传 10 个媒体文件")
	}

	group = normalizeGroup(group)
	if group == "" {
		group = "gallery"
	}
	if !isAllowedGroup(group) {
		return nil, errs.Validation("媒体分组无效")
	}
	if (group == "header" || group == "bgm") && len(files) > 1 {
		return nil, errs.Validation(group + " 分组一次只能上传 1 个文件")
	}

	total, err := s.repo.CountByPostcardID(postcardID)
	if err != nil {
		return nil, errs.Internal("获取媒体失败")
	}
	if total+len(files) > maxMediasPerCard {
		return nil, errs.Validation("单张明信片最多允许 20 个媒体文件")
	}

	if group == "header" || group == "bgm" {
		count, err := s.repo.CountByPostcardIDAndGroup(postcardID, group)
		if err != nil {
			return nil, errs.Internal("获取媒体失败")
		}
		if count > 0 {
			return nil, errs.Validation(group + " 分组仅允许 1 个媒体文件")
		}
	}

	startPosition := total + 1
	uploaded := make([]mediadomain.Media, 0, len(files))
	for index, file := range files {
		media, err := s.uploadOne(postcardID, file, mediaType, group, startPosition+index)
		if err != nil {
			return nil, err
		}
		uploaded = append(uploaded, *media)
	}
	s.enqueueAnalysis(postcardID)
	return uploaded, nil
}

func (s *Service) List(postcardID uint) ([]mediadomain.Media, error) {
	medias, err := s.repo.ListByPostcardID(postcardID)
	if err != nil {
		return nil, errs.Internal("获取媒体失败")
	}
	return medias, nil
}

func (s *Service) Reorder(postcardID uint, mediaIDs []uint) error {
	if len(mediaIDs) == 0 {
		return errs.Validation("媒体排序不能为空")
	}
	all, err := s.repo.ListByPostcardID(postcardID)
	if err != nil {
		return errs.Internal("获取媒体失败")
	}
	if len(all) != len(mediaIDs) {
		return errs.Validation("必须传入全部媒体ID进行排序")
	}

	seen := map[uint]bool{}
	for _, id := range mediaIDs {
		if seen[id] {
			return errs.Validation("必须传入全部媒体ID进行排序")
		}
		seen[id] = true
	}
	for _, media := range all {
		if !seen[media.ID] {
			return errs.Validation("必须传入全部媒体ID进行排序")
		}
	}

	if err := s.repo.Reorder(postcardID, mediaIDs); err != nil {
		return errs.Internal("更新媒体排序失败")
	}
	s.enqueueAnalysis(postcardID)
	return nil
}

func (s *Service) Delete(postcardID, mediaID uint) error {
	if err := s.repo.Delete(postcardID, mediaID); err != nil {
		return errs.NotFound("媒体不存在")
	}
	s.enqueueAnalysis(postcardID)
	return nil
}

func (s *Service) enqueueAnalysis(postcardID uint) {
	if s.ai == nil || postcardID == 0 {
		return
	}
	_, _ = s.ai.EnqueuePostcardAnalysis(context.Background(), postcardaiapp.EnqueueInput{
		PostcardID:  postcardID,
		Reason:      postcardaiapp.EnqueueReasonMediaChange,
		RequestedBy: "system",
	})
}

func validateGroupCompatibility(mediaType, group string) error {
	switch group {
	case "header":
		if mediaType != "image" {
			return errs.Validation("header 分组仅支持图片")
		}
	case "bgm":
		if mediaType != "audio" {
			return errs.Validation("bgm 分组仅支持音频")
		}
	}
	return nil
}

func normalizeGroup(group string) string {
	return strings.ToLower(strings.TrimSpace(group))
}

func isAllowedGroup(group string) bool {
	switch group {
	case "header", "gallery", "bgm":
		return true
	default:
		return false
	}
}

func normalizeType(mediaType string) string {
	return strings.ToLower(strings.TrimSpace(mediaType))
}

func detectType(filename string, explicit string) (string, error) {
	mediaType := normalizeType(explicit)
	if mediaType != "" {
		switch mediaType {
		case "image", "video", "audio":
			return mediaType, nil
		default:
			return "", errs.Validation("媒体类型无效")
		}
	}

	switch strings.ToLower(filepath.Ext(filename)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return "image", nil
	case ".mp3", ".wav", ".ogg":
		return "audio", nil
	case ".mp4", ".mov", ".avi":
		return "video", nil
	default:
		return "", errs.Validation("媒体类型无效")
	}
}

func (s *Service) uploadOne(postcardID uint, file *multipart.FileHeader, mediaType, group string, position int) (*mediadomain.Media, error) {
	if file == nil {
		return nil, errs.Validation("媒体文件不能为空")
	}

	actualType, err := detectType(file.Filename, mediaType)
	if err != nil {
		return nil, err
	}
	if err := validateGroupCompatibility(actualType, group); err != nil {
		return nil, err
	}

	src, err := file.Open()
	if err != nil {
		return nil, errs.Validation("读取媒体失败")
	}
	defer src.Close()

	data := make([]byte, file.Size)
	n, err := src.Read(data)
	if err != nil && n == 0 {
		return nil, errs.Validation("读取媒体失败")
	}
	data = data[:n]

	key := fmt.Sprintf("postcards/%d/%d-%s", postcardID, time.Now().UnixNano(), filepath.Base(file.Filename))
	url, err := s.storage.Upload(key, file.Filename, data, file.Header.Get("Content-Type"))
	if err != nil {
		return nil, errs.Internal("上传媒体失败")
	}

	width, height := 0, 0
	if actualType == "image" {
		width, height = s.processor.ImageMetadata(file.Filename, data)
	}

	media, err := s.repo.Create(&mediadomain.Media{
		PostcardID:     postcardID,
		Type:           actualType,
		URL:            url,
		StorageKey:     key,
		OriginalWidth:  width,
		OriginalHeight: height,
		FileSize:       file.Size,
		Position:       position,
		Group:          group,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})
	if err != nil {
		return nil, errs.Internal("保存媒体信息失败")
	}
	return media, nil
}

type noopStorage struct{}

func (noopStorage) Upload(key, filename string, data []byte, contentType string) (string, error) {
	return "https://cdn.example.com/" + key, nil
}

func (noopStorage) Delete(string) error { return nil }

type noopImageProcessor struct{}

func (noopImageProcessor) ImageMetadata(string, []byte) (int, int) { return 0, 0 }

type memoryRepository struct {
	nextID     uint
	mediasByID map[uint]*mediadomain.Media
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{
		nextID:     1,
		mediasByID: map[uint]*mediadomain.Media{},
	}
}

func (r *memoryRepository) Create(media *mediadomain.Media) (*mediadomain.Media, error) {
	copy := *media
	copy.ID = r.nextID
	r.nextID++
	if copy.CreatedAt.IsZero() {
		copy.CreatedAt = time.Now()
	}
	if copy.UpdatedAt.IsZero() {
		copy.UpdatedAt = copy.CreatedAt
	}
	r.mediasByID[copy.ID] = &copy
	result := copy
	return &result, nil
}

func (r *memoryRepository) ListByPostcardID(postcardID uint) ([]mediadomain.Media, error) {
	items := make([]mediadomain.Media, 0)
	for _, media := range r.mediasByID {
		if media.PostcardID == postcardID {
			copy := *media
			items = append(items, copy)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Position < items[j].Position
	})
	return items, nil
}

func (r *memoryRepository) CountByPostcardID(postcardID uint) (int, error) {
	count := 0
	for _, media := range r.mediasByID {
		if media.PostcardID == postcardID {
			count++
		}
	}
	return count, nil
}

func (r *memoryRepository) CountByPostcardIDAndGroup(postcardID uint, group string) (int, error) {
	count := 0
	for _, media := range r.mediasByID {
		if media.PostcardID == postcardID && media.Group == group {
			count++
		}
	}
	return count, nil
}

func (r *memoryRepository) Reorder(postcardID uint, mediaIDs []uint) error {
	for index, id := range mediaIDs {
		media, ok := r.mediasByID[id]
		if !ok || media.PostcardID != postcardID {
			return errs.Validation("必须传入全部媒体ID进行排序")
		}
		media.Position = index + 1
		media.UpdatedAt = time.Now()
	}
	return nil
}

func (r *memoryRepository) Delete(postcardID, mediaID uint) error {
	media, ok := r.mediasByID[mediaID]
	if !ok || media.PostcardID != postcardID {
		return errs.NotFound("媒体不存在")
	}
	delete(r.mediasByID, mediaID)
	return nil
}

func (r *memoryRepository) DeleteByPostcardID(postcardID uint) error {
	for id, media := range r.mediasByID {
		if media.PostcardID == postcardID {
			delete(r.mediasByID, id)
		}
	}
	return nil
}
