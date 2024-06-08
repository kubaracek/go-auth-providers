package lib

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandomTokenBoundToSize(t *testing.T) {
	t.Parallel()
	token, err := RandomToken(15)
	assert.NoError(t, err)
	assert.Len(t, token, 32)

	token2, err := RandomToken(40)
	assert.NoError(t, err)
	assert.Len(t, token2, 40)

	token3, err := RandomToken(100)
	assert.NoError(t, err)
	assert.Len(t, token3, 64)
}
