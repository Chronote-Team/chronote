package app

import (
	"strings"

	usersdomain "chronote-refactor/internal/modules/users/domain"
	"chronote-refactor/internal/shared/errs"
)

type UserRepository interface {
	FindByEmail(email string) (*usersdomain.User, error)
	FindByID(id uint) (*usersdomain.User, error)
}

type PasswordVerifier interface {
	Verify(string, string) (bool, error)
}

type TokenClaims struct {
	UserID    uint
	Name      string
	TokenType string
}

type TokenService interface {
	GenerateTokenPair(userID uint, username string) (string, string, error)
	ParseToken(token string) (*TokenClaims, error)
	AccessExpirySeconds() int64
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type Service struct {
	repo      UserRepository
	passwords PasswordVerifier
	tokens    TokenService
	blacklist TokenBlacklist
}

func NewService(repo UserRepository, passwords PasswordVerifier, tokens TokenService) *Service {
	return &Service{
		repo:      repo,
		passwords: passwords,
		tokens:    tokens,
		blacklist: NewMemoryBlacklist(),
	}
}

func (s *Service) SetBlacklist(blacklist TokenBlacklist) {
	if blacklist != nil {
		s.blacklist = blacklist
	}
}

func (s *Service) ValidateTokenType(tokenType string) error {
	if tokenType != "access" {
		return errs.Unauthorized("Need Use Aceess Token to Authorize!")
	}
	return nil
}

func (s *Service) Login(email, password string) (*Tokens, error) {
	user, err := s.repo.FindByEmail(strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return nil, errs.Internal("登录失败")
	}
	if user == nil {
		return nil, errs.Unauthorized("用户不存在或密码错误")
	}
	ok, err := s.passwords.Verify(password, user.PasswordHash)
	if err != nil || !ok {
		return nil, errs.Unauthorized("用户不存在或密码错误")
	}
	accessToken, refreshToken, err := s.tokens.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, errs.Internal("token 生成失败")
	}
	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.tokens.AccessExpirySeconds(),
	}, nil
}

func (s *Service) RefreshToken(refreshToken string) (*Tokens, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, errs.Validation("refresh_token 不能为空")
	}
	blacklisted, err := s.blacklist.IsBlacklisted(refreshToken)
	if err != nil {
		return nil, errs.Internal("token 验证失败")
	}
	if blacklisted {
		return nil, errs.Unauthorized("refresh token 已被撤销")
	}
	claims, err := s.tokens.ParseToken(refreshToken)
	if err != nil {
		return nil, errs.Unauthorized("refresh token 无效或已过期")
	}
	if claims.TokenType != "refresh" {
		return nil, errs.Unauthorized("需要使用 refresh token")
	}
	user, err := s.repo.FindByID(claims.UserID)
	if err != nil || user == nil {
		return nil, errs.Unauthorized("用户不存在")
	}
	accessToken, newRefreshToken, err := s.tokens.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, errs.Internal("token 生成失败")
	}
	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    s.tokens.AccessExpirySeconds(),
	}, nil
}

func (s *Service) Logout(userID uint, accessToken, refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return errs.Validation("refresh_token 不能为空")
	}
	claims, err := s.tokens.ParseToken(refreshToken)
	if err != nil {
		return errs.Unauthorized("refresh token 无效或已过期")
	}
	if claims.TokenType != "refresh" {
		return errs.Unauthorized("需要使用 refresh token")
	}
	if claims.UserID != userID {
		return errs.Unauthorized("refresh token 不属于当前用户")
	}
	if err := s.blacklist.BlacklistTokenPair(accessToken, refreshToken); err != nil {
		return errs.Internal("登出失败")
	}
	return nil
}
