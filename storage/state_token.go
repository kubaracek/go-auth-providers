package storage

// StateTokenStore interface for storing and burning state tokens
// state token is a one-time use token used to validate response from 3rd party providers (e.g. OAuth2) really comes from them.
// It shouldn't take user much time to authenticate, so the state token should have a TTL and be valid for a short period of time
// (e.g. 5 minutes).
type StateTokenStore interface {
	// StoreStateToken stores the token for the authentication
	// subsequent calls to BurnToken with the same token should invalidate the token
	StoreStateToken(token, session, provider string) error
	// BurnStateToken invalidates the token used for the authentication
	// returns the given session and provider. An error if the token is not found or any other error occurs.
	// subsequent calls to BurnStateToken with the same token should return an error
	// BurnStateToken should also check if the token is still valid and not expired
	BurnStateToken(token string) (session, provider string, err error)
}
