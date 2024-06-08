package external

import (
	"context"
	"encoding/json"
	"github.com/kubaracek/go-auth-providers"
	"github.com/kubaracek/go-auth-providers/storage"
	"github.com/markbates/goth"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"net/url"
	"strings"
	"testing"
	"time"
)

func MockProviderForTest(validCode string) RegisterProviderFn {
	return func() (string, Provider) {
		return "mock", &MockProvider{
			validCode: validCode,
		}
	}
}

type MockSession struct {
	AuthURL   string
	ValidCode string
}

// GetAuthURL will return the URL set by calling the `BeginAuth` function on the GitHub provider.
func (s MockSession) GetAuthURL() (string, error) {
	if s.AuthURL == "" {
		return "", errors.New(goth.NoAuthUrlErrorMessage)
	}
	return s.AuthURL, nil
}

// Authorize the session with GitHub and return the access token to be stored for future use.
func (s MockSession) Authorize(provider goth.Provider, params goth.Params) (string, error) {
	code := params.Get("code")
	if code == s.ValidCode {
		return "token", nil
	}

	return "", errors.New("invalid token received from provider")
}

// Marshal the session into a string
func (s MockSession) Marshal() string {
	b, _ := json.Marshal(s)
	return string(b)
}

func (s MockSession) String() string {
	return s.Marshal()
}

// UnmarshalSession will unmarshal a JSON string into a session.
func (p *MockProvider) UnmarshalSession(data string) (goth.Session, error) {
	sess := &MockSession{}
	err := json.NewDecoder(strings.NewReader(data)).Decode(sess)
	return sess, err
}

type MockProvider struct {
	validCode string
}

func (p *MockProvider) BeginAuth(state string) (goth.Session, error) {
	url := "http://localhost:8080/auth/mock/callback?state=" + state
	return &MockSession{
		AuthURL:   url,
		ValidCode: p.validCode,
	}, nil
}

func (p *MockProvider) Name() string {
	return "mock"
}

func (p *MockProvider) Debug(debug bool) {}

func (p *MockProvider) FetchUser(session goth.Session) (goth.User, error) {
	return goth.User{}, nil
}

func (p *MockProvider) RefreshToken(refreshToken string) (*oauth2.Token, error) {
	return nil, nil
}

func (p *MockProvider) RefreshTokenAvailable() bool {
	return false
}

func (p *MockProvider) SetName(name string) {}

func TestRegisterProvider(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	stateTokenStore := storage.NewInMemStateTokenStore(ctx, 1*time.Minute, 1*time.Minute)
	manager := NewProviderManager(stateTokenStore, MockProviderForTest("key"))

	assert.NotEmpty(t, manager.providers)
	assert.Len(t, manager.providers, 1)

	provider := manager.getProvider("mock")
	assert.NotNil(t, provider)
}

func TestBeginAuth(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	stateTokenStore := storage.NewInMemStateTokenStore(ctx, 1*time.Minute, 1*time.Minute)
	manager := NewProviderManager(stateTokenStore, MockProviderForTest("code"))

	authUrl, err := manager.BeginAuth("mock")
	assert.NoError(t, err)
	assert.NotEmpty(t, authUrl)
}

func TestCompleteAuth(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	stateTokenStore := storage.NewInMemStateTokenStore(ctx, 1*time.Minute, 1*time.Minute)
	manager := NewProviderManager(stateTokenStore, MockProviderForTest("code"))

	authUrl, err := manager.BeginAuth("mock")
	assert.NoError(t, err)
	assert.NotEmpty(t, authUrl)

	stateToken, err := parseStateTokenFromUrl(authUrl)
	assert.NoError(t, err)

	authData := lib.MapOfStrings(map[string]string{"code": "code"})

	user, err := manager.CompleteAuth(stateToken, authData)
	assert.NoError(t, err)

	assert.NotNil(t, user)
}

func parseStateTokenFromUrl(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	queryParams := parsedURL.Query()

	return queryParams.Get("state"), nil
}
