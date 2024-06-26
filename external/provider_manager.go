package external

import (
	"github.com/kubaracek/go-auth-providers"
	"github.com/kubaracek/go-auth-providers/storage"
	"github.com/markbates/goth"
	"sync"
)

// Provider wrapper for goth.Provider
type Provider goth.Provider

// RegisterProviderFn Even though the Provider has the Name() method, we rather give them a name here
// as it's not given that the underlying provider will implement them. It's a simple precaution.
type RegisterProviderFn func() (providerName string, provider Provider)

type ProviderManager interface {
	BeginAuth(providerName string) (authUrl string, err error)
	CompleteAuth(stateToken string, authData lib.Params) (lib.User, error)
}

type ProviderManagerImpl struct {
	stateTokenStore storage.StateTokenStore
	mu              sync.RWMutex
	providers       map[string]Provider
}

func NewProviderManager(stateTokenStore storage.StateTokenStore, opts ...RegisterProviderFn) *ProviderManagerImpl {
	pm := &ProviderManagerImpl{
		stateTokenStore: stateTokenStore,
		providers:       make(map[string]Provider),
	}

	for _, opt := range opts {
		name, provider := opt()
		pm.registerProvider(name, provider)
	}

	return pm
}

func (pm *ProviderManagerImpl) BeginAuth(providerName string) (string, error) {
	provider := pm.getProvider(providerName)
	if provider == nil {
		return "", lib.InternalError{Message: "provider not found"}
	}

	stateToken, err := lib.RandomToken(64)
	if err != nil {
		return "", lib.InternalError{Message: "failed to generate state token"}
	}

	session, err := provider.BeginAuth(stateToken)
	if err != nil {
		return "", lib.InternalError{Message: "failed to begin auth"}
	}

	err = pm.stateTokenStore.StoreStateToken(stateToken, session.Marshal(), providerName)
	if err != nil {
		return "", lib.InternalError{Message: "failed to store state token"}
	}

	authUrl, err := session.GetAuthURL()
	if err != nil {
		return "", lib.InternalError{Message: "failed to get auth url"}
	}

	return authUrl, nil
}

func (pm *ProviderManagerImpl) CompleteAuth(stateToken string, authData lib.Params) (lib.User, error) {
	var user lib.User
	sessionData, providerName, err := pm.stateTokenStore.BurnStateToken(stateToken)
	if err != nil {
		return user, lib.InternalError{Message: "failed to burn state token"}
	}

	provider := pm.getProvider(providerName)
	if provider == nil {
		return user, lib.InternalError{Message: "provider not found"}
	}

	session, err := provider.UnmarshalSession(sessionData)
	if err != nil {
		return user, lib.InternalError{Message: "failed to unmarshal session"}
	}

	_, err = session.Authorize(goth.Provider(provider), authData)
	if err != nil {
		return user, lib.UserAuthError{Message: "failed to authorize"}
	}

	u, err := provider.FetchUser(session)
	if err != nil {
		return user, lib.UserAuthError{Message: "failed to fetch user"}
	}

	user = lib.User(u)

	return user, nil
}

func (pm *ProviderManagerImpl) registerProvider(name string, p Provider) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.providers[name] = p
}

func (pm *ProviderManagerImpl) getProvider(name string) Provider {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.providers[name]
}
