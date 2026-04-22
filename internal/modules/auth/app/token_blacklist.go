package app

import "sync"

type TokenBlacklist interface {
	IsBlacklisted(token string) (bool, error)
	BlacklistTokenPair(accessToken, refreshToken string) error
}

type MemoryBlacklist struct {
	mu     sync.RWMutex
	tokens map[string]struct{}
}

func NewMemoryBlacklist() *MemoryBlacklist {
	return &MemoryBlacklist{tokens: map[string]struct{}{}}
}

func (b *MemoryBlacklist) IsBlacklisted(token string) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, ok := b.tokens[token]
	return ok, nil
}

func (b *MemoryBlacklist) BlacklistTokenPair(accessToken, refreshToken string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if accessToken != "" {
		b.tokens[accessToken] = struct{}{}
	}
	if refreshToken != "" {
		b.tokens[refreshToken] = struct{}{}
	}
	return nil
}
