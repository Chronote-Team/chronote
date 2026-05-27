package app

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	usersdomain "chronote-refactor/internal/modules/users/domain"
	"chronote-refactor/internal/shared/errs"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

const (
	minPasswordLength    = 6
	maxPasswordLength    = 72
	maxUsernameLength    = 50
	maxDisplayNameLength = 100
	maxEmailLength       = 255
	maxAvatarSize        = 10 * 1024 * 1024
)

type RegisterInput struct {
	Username    string
	DisplayName string
	Email       string
	Password    string
}

type Service struct {
	repo          Repository
	passwords     PasswordManager
	avatarStorage AvatarStorage
}

func NewService(repo Repository, passwords PasswordManager) *Service {
	if repo == nil {
		repo = newMemoryRepository()
	}
	if passwords == nil {
		passwords = fallbackPasswordManager{}
	}
	return &Service{repo: repo, passwords: passwords, avatarStorage: fallbackAvatarStorage{}}
}

func (s *Service) Repository() Repository {
	return s.repo
}

func (s *Service) SetAvatarStorage(storage AvatarStorage) {
	if storage == nil {
		storage = fallbackAvatarStorage{}
	}
	s.avatarStorage = storage
}

func (s *Service) Register(input RegisterInput) (*usersdomain.User, error) {
	username := strings.TrimSpace(input.Username)
	displayName := strings.TrimSpace(input.DisplayName)
	email := strings.ToLower(strings.TrimSpace(input.Email))
	password := input.Password

	if username == "" {
		return nil, errs.Validation("username 不能为空")
	}
	if len(username) > maxUsernameLength {
		return nil, errs.Validation("username 长度不能超过 50 个字符")
	}
	if !usernamePattern.MatchString(username) {
		return nil, errs.Validation("username 只能包含字母、数字和下划线")
	}
	if displayName != "" && len(displayName) > maxDisplayNameLength {
		return nil, errs.Validation("display_name 长度不能超过 100 个字符")
	}
	if email == "" {
		return nil, errs.Validation("email 不能为空")
	}
	if len(email) > maxEmailLength {
		return nil, errs.Validation("email 长度不能超过 255 个字符")
	}
	if strings.TrimSpace(password) == "" {
		return nil, errs.Validation("password 不能为空")
	}
	if len(password) < minPasswordLength || len(password) > maxPasswordLength {
		return nil, errs.Validation("password 长度必须在 6 到 72 个字符之间")
	}
	if displayName == "" {
		displayName = username
	}

	if existing, err := s.repo.FindByUsername(username); err != nil {
		return nil, errs.Internal("用户注册失败")
	} else if existing != nil {
		return nil, errs.Conflict("username 已存在")
	}
	if existing, err := s.repo.FindByEmail(email); err != nil {
		return nil, errs.Internal("用户注册失败")
	} else if existing != nil {
		return nil, errs.Conflict("email 已被使用")
	}

	hashedPassword, err := s.passwords.Hash(password)
	if err != nil {
		return nil, errs.Internal("密码加密失败")
	}

	user := &usersdomain.User{
		Username:     username,
		DisplayName:  displayName,
		Email:        email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := s.repo.Create(user); err != nil {
		return nil, errs.Internal("用户注册失败")
	}

	return user, nil
}

func (s *Service) GetUserInfo(userID uint) (*usersdomain.User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, errs.Internal("获取用户信息失败")
	}
	if user == nil {
		return nil, errs.NotFound("用户不存在")
	}
	return user, nil
}

func (s *Service) UpdateDisplayName(userID uint, displayName string) error {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return errs.Validation("display_name 不能为空")
	}
	if len(displayName) > maxDisplayNameLength {
		return errs.Validation("display_name 长度不能超过 100 个字符")
	}
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return errs.Internal("更新显示名称失败")
	}
	if user == nil {
		return errs.NotFound("用户不存在")
	}
	user.DisplayName = displayName
	user.UpdatedAt = time.Now()
	if err := s.repo.Update(user); err != nil {
		return errs.Internal("更新显示名称失败")
	}
	return nil
}

func (s *Service) UpdatePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return errs.Internal("更新密码失败")
	}
	if user == nil {
		return errs.NotFound("用户不存在")
	}
	if strings.TrimSpace(newPassword) == "" {
		return errs.Validation("new_password 不能为空")
	}
	if len(newPassword) < minPasswordLength || len(newPassword) > maxPasswordLength {
		return errs.Validation("new_password 长度必须在 6 到 72 个字符之间")
	}
	matched, err := s.passwords.Verify(oldPassword, user.PasswordHash)
	if err != nil || !matched {
		return errs.Unauthorized("旧密码错误")
	}
	hashedPassword, err := s.passwords.Hash(newPassword)
	if err != nil {
		return errs.Internal("更新密码失败")
	}
	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()
	if err := s.repo.Update(user); err != nil {
		return errs.Internal("更新密码失败")
	}
	return nil
}

func (s *Service) UpdateAvatar(userID uint, avatarURL string) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return errs.Internal("更新用户头像失败")
	}
	if user == nil {
		return errs.NotFound("用户不存在")
	}
	user.Avatar = avatarURL
	user.UpdatedAt = time.Now()
	if err := s.repo.Update(user); err != nil {
		return errs.Internal("更新用户头像失败")
	}
	return nil
}

func (s *Service) UploadAvatar(userID uint, file *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", errs.Validation("请上传头像文件")
	}
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return "", errs.Internal("更新用户头像失败")
	}
	if user == nil {
		return "", errs.NotFound("用户不存在")
	}
	filename := filepath.Base(file.Filename)
	contentType := strings.ToLower(strings.TrimSpace(file.Header.Get("Content-Type")))
	if err := validateAvatarFile(filename, contentType, file.Size); err != nil {
		return "", err
	}

	src, err := file.Open()
	if err != nil {
		return "", errs.Validation("读取头像失败")
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return "", errs.Validation("读取头像失败")
	}

	key := fmt.Sprintf("avatars/%d/%d-%s", userID, time.Now().UnixNano(), filename)
	url, err := s.avatarStorage.Upload(key, filename, data, contentType)
	if err != nil {
		return "", errs.Internal("上传头像失败")
	}

	user.Avatar = url
	user.UpdatedAt = time.Now()
	if err := s.repo.Update(user); err != nil {
		return "", errs.Internal("更新用户头像失败")
	}
	return url, nil
}

func validateAvatarFile(filename, contentType string, size int64) error {
	if filename == "." || filename == string(filepath.Separator) || strings.TrimSpace(filename) == "" {
		return errs.Validation("头像文件类型无效")
	}
	if size <= 0 {
		return errs.Validation("头像文件不能为空")
	}
	if size > maxAvatarSize {
		return errs.Validation("头像文件大小不能超过 10MB")
	}
	extension := strings.ToLower(filepath.Ext(filename))
	switch extension {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
	default:
		return errs.Validation("头像文件类型无效")
	}
	if contentType != "" && contentType != "application/octet-stream" && !strings.HasPrefix(contentType, "image/") {
		return errs.Validation("头像文件类型无效")
	}
	return nil
}

type AvatarStorage interface {
	Upload(key, filename string, data []byte, contentType string) (string, error)
}

type fallbackAvatarStorage struct{}

func (fallbackAvatarStorage) Upload(key, filename string, data []byte, contentType string) (string, error) {
	return "https://media.example.com/" + key, nil
}

type fallbackPasswordManager struct{}

func (fallbackPasswordManager) Hash(password string) (string, error) {
	return "hashed:" + password, nil
}

func (fallbackPasswordManager) Verify(plain, stored string) (bool, error) {
	if !strings.HasPrefix(stored, "hashed:") {
		return false, errors.New("invalid stored password")
	}
	return stored == "hashed:"+plain, nil
}

type memoryRepository struct {
	nextID    uint
	usersByID map[uint]*usersdomain.User
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{
		nextID:    1,
		usersByID: map[uint]*usersdomain.User{},
	}
}

func (r *memoryRepository) Create(user *usersdomain.User) error {
	user.ID = r.nextID
	r.nextID++
	copy := *user
	r.usersByID[user.ID] = &copy
	return nil
}

func (r *memoryRepository) FindByID(id uint) (*usersdomain.User, error) {
	user, ok := r.usersByID[id]
	if !ok {
		return nil, nil
	}
	copy := *user
	return &copy, nil
}

func (r *memoryRepository) FindByEmail(email string) (*usersdomain.User, error) {
	for _, user := range r.usersByID {
		if user.Email == email {
			copy := *user
			return &copy, nil
		}
	}
	return nil, nil
}

func (r *memoryRepository) FindByUsername(username string) (*usersdomain.User, error) {
	for _, user := range r.usersByID {
		if user.Username == username {
			copy := *user
			return &copy, nil
		}
	}
	return nil, nil
}

func (r *memoryRepository) Update(user *usersdomain.User) error {
	copy := *user
	r.usersByID[user.ID] = &copy
	return nil
}
