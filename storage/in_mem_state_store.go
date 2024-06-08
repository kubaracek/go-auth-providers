package storage

import (
	"context"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type state struct {
	session  string
	provider string
	ttl      time.Time
}

type InMemAuthTokenStoreImpl struct {
	ttl    time.Duration
	tokens map[string]state
	mu     sync.RWMutex
}

// NewInMemStateTokenStore creates a new in-memory state token store
// ttl is the time-to-live for the tokens
// expired tokens are automatically removed from the store using a goroutine.
// You can't use this store for a multi-pod setup, as the tokens are stored in memory only in the current pod.
func NewInMemStateTokenStore(ctx context.Context, ttl time.Duration, cleanExpiredTokensEvery time.Duration) *InMemAuthTokenStoreImpl {
	store := &InMemAuthTokenStoreImpl{
		ttl:    ttl,
		tokens: make(map[string]state),
	}

	go func() {
		ticker := time.NewTicker(cleanExpiredTokensEvery)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				store.mu.Lock()
				now := time.Now()
				for token, _state := range store.tokens {
					if now.After(_state.ttl) {
						delete(store.tokens, token)
					}
				}
				store.mu.Unlock()
			}
		}
	}()

	return store
}

func (s *InMemAuthTokenStoreImpl) StoreStateToken(token, session, provider string) error {
	expires := time.Now().Add(s.ttl)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token] = state{
		session:  session,
		provider: provider,
		ttl:      expires,
	}
	return nil
}

func (s *InMemAuthTokenStoreImpl) BurnStateToken(token string) (session, provider string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_state, ok := s.tokens[token]
	if !ok {
		return "", "", errors.New("auth token not found")
	}

	if time.Now().After(_state.ttl) {
		delete(s.tokens, token)
		return "", "", errors.New("auth token expired")
	}
	delete(s.tokens, token)
	return _state.session, _state.provider, nil
}
