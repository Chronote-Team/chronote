package app

import (
	"context"
	"encoding/json"
	"math/rand"
	"sort"
	"strings"
	"time"

	mediadomain "chronote-refactor/internal/modules/media/domain"
	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	postcardsdomain "chronote-refactor/internal/modules/postcards/domain"
	"chronote-refactor/internal/shared/errs"
	sharedpagination "chronote-refactor/internal/shared/pagination"
)

const (
	maxTitleLength  = 200
	maxContentBytes = 64 * 1024
)

type ListInput struct {
	Page       int
	PageSize   int
	Visibility string
	SortBy     string
	Order      string
}

type CreateInput struct {
	Title      string
	Content    json.RawMessage
	Visibility string
}

type UpdateInput struct {
	Title      *string
	Content    *json.RawMessage
	Visibility *string
}

type Service struct {
	repo    Repository
	authors AuthorRepository
	medias  MediaRepository
	ai      AnalysisEnqueuer
}

func NewService(repo Repository, authors AuthorRepository, medias MediaRepository) *Service {
	if repo == nil {
		repo = newMemoryRepository()
	}
	if medias == nil {
		medias = noopMediaRepository{}
	}
	return &Service{repo: repo, authors: authors, medias: medias, ai: postcardaiapp.NoopEnqueuer{}}
}

func (s *Service) SetAnalysisEnqueuer(enqueuer AnalysisEnqueuer) {
	if enqueuer == nil {
		enqueuer = postcardaiapp.NoopEnqueuer{}
	}
	s.ai = enqueuer
}

func (s *Service) Create(userID uint, input CreateInput) (*postcardsdomain.Postcard, error) {
	visibility := normalizeVisibility(input.Visibility)
	if visibility == "" {
		visibility = "private"
	}
	if err := validateVisibility(visibility); err != nil {
		return nil, err
	}

	title, err := validateTitle(input.Title)
	if err != nil {
		return nil, err
	}
	content, err := validateContent(input.Content)
	if err != nil {
		return nil, err
	}

	postcard := &postcardsdomain.Postcard{
		Title:      title,
		Content:    content,
		Visibility: visibility,
		AuthorID:   userID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := s.repo.Create(postcard); err != nil {
		return nil, errs.Internal("创建明信片失败")
	}
	s.enqueueAnalysis(postcard.ID, postcardaiapp.EnqueueReasonCreate)
	return s.attachRelations(postcard)
}

func (s *Service) List(userID uint, input ListInput) ([]postcardsdomain.Postcard, sharedpagination.Page, error) {
	page, pageSize := sharedpagination.Normalize(input.Page, input.PageSize)

	all, err := s.repo.List()
	if err != nil {
		return nil, sharedpagination.Page{}, errs.Internal("获取明信片列表失败")
	}

	visibility := normalizeVisibility(input.Visibility)
	if input.Visibility != "" && visibility == "" {
		return nil, sharedpagination.Page{}, errs.Validation("visibility 无效")
	}

	filtered := make([]postcardsdomain.Postcard, 0, len(all))
	for _, postcard := range all {
		if !canAccess(userID, postcard.AuthorID, postcard.Visibility) {
			continue
		}
		if visibility != "" && postcard.Visibility != visibility {
			continue
		}
		filtered = append(filtered, postcard)
	}

	sortBy, order := normalizeSort(input.SortBy, input.Order)
	sort.Slice(filtered, func(i, j int) bool {
		var less bool
		if sortBy == "updated_at" {
			less = filtered[i].UpdatedAt.Before(filtered[j].UpdatedAt)
		} else {
			less = filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
		}
		if order == "asc" {
			return less
		}
		return !less
	})

	total := int64(len(filtered))
	start := (page - 1) * pageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	items := make([]postcardsdomain.Postcard, 0, end-start)
	for i := start; i < end; i++ {
		postcard, err := s.attachRelations(&filtered[i])
		if err != nil {
			return nil, sharedpagination.Page{}, errs.Internal("获取明信片列表失败")
		}
		items = append(items, *postcard)
	}

	return items, sharedpagination.Page{Page: page, PageSize: pageSize, Total: total}, nil
}

func (s *Service) GetDetail(userID, postcardID uint) (*postcardsdomain.Postcard, error) {
	postcard, err := s.repo.FindByID(postcardID)
	if err != nil {
		return nil, errs.Internal("获取明信片详情失败")
	}
	if postcard == nil {
		return nil, errs.NotFound("明信片不存在")
	}
	if !canAccess(userID, postcard.AuthorID, postcard.Visibility) {
		return nil, errs.Forbidden("无权限访问该明信片")
	}
	return s.attachRelations(postcard)
}

func (s *Service) GetRandom(userID uint) (*postcardsdomain.Postcard, error) {
	postcard, err := s.repo.FindRandomAccessible(userID)
	if err != nil {
		return nil, errs.Internal("获取随机明信片失败")
	}
	if postcard == nil {
		return nil, errs.NotFound("明信片不存在")
	}
	postcard, err = s.attachRelations(postcard)
	if err != nil {
		return nil, errs.Internal("获取随机明信片失败")
	}
	return postcard, nil
}

func (s *Service) EnsureOwner(userID, postcardID uint) (*postcardsdomain.Postcard, error) {
	postcard, err := s.repo.FindByID(postcardID)
	if err != nil {
		return nil, errs.Internal("获取明信片失败")
	}
	if postcard == nil {
		return nil, errs.NotFound("明信片不存在")
	}
	if postcard.AuthorID != userID {
		return nil, errs.Forbidden("无权限操作该明信片")
	}
	return s.attachRelations(postcard)
}

func (s *Service) Update(userID, postcardID uint, input UpdateInput) error {
	postcard, err := s.repo.FindByID(postcardID)
	if err != nil {
		return errs.Internal("更新明信片失败")
	}
	if postcard == nil {
		return errs.NotFound("明信片不存在")
	}
	if postcard.AuthorID != userID {
		return errs.Forbidden("无权限操作该明信片")
	}

	updated := false
	if input.Title != nil {
		title, err := validateTitle(*input.Title)
		if err != nil {
			return err
		}
		postcard.Title = title
		updated = true
	}
	if input.Content != nil {
		content, err := validateContent(*input.Content)
		if err != nil {
			return err
		}
		postcard.Content = content
		updated = true
	}
	if input.Visibility != nil {
		visibility := normalizeVisibility(*input.Visibility)
		if visibility == "" {
			return errs.Validation("visibility 无效")
		}
		if err := validateVisibility(visibility); err != nil {
			return err
		}
		postcard.Visibility = visibility
		updated = true
	}
	if !updated {
		return errs.Validation("没有可更新的字段")
	}
	postcard.UpdatedAt = time.Now()
	if err := s.repo.Update(postcard); err != nil {
		return errs.Internal("更新明信片失败")
	}
	s.enqueueAnalysis(postcard.ID, postcardaiapp.EnqueueReasonUpdate)
	return nil
}

func (s *Service) Delete(userID, postcardID uint) error {
	postcard, err := s.repo.FindByID(postcardID)
	if err != nil {
		return errs.Internal("删除明信片失败")
	}
	if postcard == nil {
		return errs.NotFound("明信片不存在")
	}
	if postcard.AuthorID != userID {
		return errs.Forbidden("无权限操作该明信片")
	}
	if err := s.medias.DeleteByPostcardID(postcardID); err != nil {
		return errs.Internal("删除媒体失败")
	}
	if err := s.repo.Delete(postcardID); err != nil {
		return errs.Internal("删除明信片失败")
	}
	return nil
}

func (s *Service) enqueueAnalysis(postcardID uint, reason postcardaiapp.EnqueueReason) {
	if s.ai == nil || postcardID == 0 {
		return
	}
	_, _ = s.ai.EnqueuePostcardAnalysis(context.Background(), postcardaiapp.EnqueueInput{
		PostcardID:  postcardID,
		Reason:      reason,
		RequestedBy: "system",
	})
}

func validateTitle(input string) (string, error) {
	title := strings.TrimSpace(input)
	if title == "" {
		return "", errs.Validation("title 不能为空")
	}
	if len(title) > maxTitleLength {
		return "", errs.Validation("title 长度不能超过 200 个字符")
	}
	return title, nil
}

func validateContent(input json.RawMessage) (json.RawMessage, error) {
	trimmed := strings.TrimSpace(string(input))
	if trimmed == "" {
		return nil, errs.Validation("content 不能为空")
	}
	if len(trimmed) > maxContentBytes {
		return nil, errs.Validation("content 长度不能超过 65536 字节")
	}
	if !json.Valid([]byte(trimmed)) {
		return nil, errs.Validation("content 无效")
	}
	var decoded any
	if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
		return nil, errs.Validation("content 无效")
	}
	switch decoded.(type) {
	case map[string]any, []any:
	default:
		return nil, errs.Validation("content 必须是 JSON 对象或数组")
	}
	return json.RawMessage([]byte(trimmed)), nil
}

func normalizeVisibility(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validateVisibility(value string) error {
	switch value {
	case "public", "private":
		return nil
	default:
		return errs.Validation("visibility 无效")
	}
}

func normalizeSort(sortBy, order string) (string, string) {
	sortBy = strings.ToLower(strings.TrimSpace(sortBy))
	if sortBy != "updated_at" {
		sortBy = "created_at"
	}
	order = strings.ToLower(strings.TrimSpace(order))
	if order != "asc" {
		order = "desc"
	}
	return sortBy, order
}

func (s *Service) attachRelations(postcard *postcardsdomain.Postcard) (*postcardsdomain.Postcard, error) {
	copy := *postcard

	if s.authors != nil {
		author, err := s.authors.FindByID(copy.AuthorID)
		if err != nil {
			return nil, err
		}
		if author != nil {
			copy.Author = &postcardsdomain.Author{
				ID:          author.ID,
				Username:    author.Username,
				DisplayName: author.DisplayName,
				Avatar:      author.Avatar,
			}
		}
	}

	medias, err := s.medias.ListByPostcardID(copy.ID)
	if err != nil {
		return nil, err
	}
	copy.Medias = medias
	return &copy, nil
}

type noopMediaRepository struct{}

func (noopMediaRepository) ListByPostcardID(uint) ([]mediadomain.Media, error) { return nil, nil }
func (noopMediaRepository) DeleteByPostcardID(uint) error                      { return nil }

type memoryRepository struct {
	nextID        uint
	postcardsByID map[uint]*postcardsdomain.Postcard
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{
		nextID:        1,
		postcardsByID: map[uint]*postcardsdomain.Postcard{},
	}
}

func (r *memoryRepository) Create(postcard *postcardsdomain.Postcard) error {
	copy := *postcard
	copy.ID = r.nextID
	r.nextID++
	r.postcardsByID[copy.ID] = &copy
	postcard.ID = copy.ID
	return nil
}

func (r *memoryRepository) FindByID(id uint) (*postcardsdomain.Postcard, error) {
	postcard, ok := r.postcardsByID[id]
	if !ok {
		return nil, nil
	}
	copy := *postcard
	return &copy, nil
}

func (r *memoryRepository) FindRandomAccessible(userID uint) (*postcardsdomain.Postcard, error) {
	candidates := make([]*postcardsdomain.Postcard, 0, len(r.postcardsByID))
	for _, postcard := range r.postcardsByID {
		if canAccess(userID, postcard.AuthorID, postcard.Visibility) {
			candidates = append(candidates, postcard)
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	copy := *candidates[rand.Intn(len(candidates))]
	return &copy, nil
}

func (r *memoryRepository) List() ([]postcardsdomain.Postcard, error) {
	items := make([]postcardsdomain.Postcard, 0, len(r.postcardsByID))
	for _, postcard := range r.postcardsByID {
		copy := *postcard
		items = append(items, copy)
	}
	return items, nil
}

func (r *memoryRepository) Update(postcard *postcardsdomain.Postcard) error {
	copy := *postcard
	r.postcardsByID[copy.ID] = &copy
	return nil
}

func (r *memoryRepository) Delete(id uint) error {
	delete(r.postcardsByID, id)
	return nil
}
