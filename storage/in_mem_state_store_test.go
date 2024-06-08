package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMemoryAuthTokenStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("store and burn token", func(t *testing.T) {
		t.Parallel()
		storage := NewInMemStateTokenStore(ctx, 1*time.Minute, 1*time.Minute)

		assert.NoError(t, storage.StoreStateToken("a-token", "a-session", "a-provider"))

		session, provider, err := storage.BurnStateToken("a-token")
		assert.NoError(t, err)

		assert.Equal(t, "a-session", session)
		assert.Equal(t, "a-provider", provider)

		// burn the same token again
		_, _, err = storage.BurnStateToken("a-token")
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("burn expired token", func(t *testing.T) {
		t.Parallel()
		// negative duration as tll should make new tokens expire immediately
		storage := NewInMemStateTokenStore(ctx, -1*time.Minute, 1*time.Minute)

		assert.NoError(t, storage.StoreStateToken("b-token", "b-session", "b-provider"))

		_, _, err := storage.BurnStateToken("b-token")
		assert.ErrorContains(t, err, "expired")
		_, _, err = storage.BurnStateToken("b-token")
		assert.ErrorContains(t, err, "not found")
	})

	t.Run("clean expired tokens", func(t *testing.T) {
		t.Parallel()
		storage := NewInMemStateTokenStore(ctx, 1*time.Millisecond, 1*time.Millisecond)

		assert.NoError(t, storage.StoreStateToken("c-token", "c-session", "c-provider"))

		time.Sleep(10 * time.Millisecond)

		_, _, err := storage.BurnStateToken("c-token")

		assert.ErrorContains(t, err, "not found")
	})
}

func BenchmarkInMemoryStore(b *testing.B) {
	ctx := context.Background()
	storage := NewInMemStateTokenStore(ctx, 1*time.Minute, 1*time.Minute)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.StoreStateToken("token-"+string(rune(i)), "session", "provider")
	}

	b.StopTimer()

}
