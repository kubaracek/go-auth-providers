package lib

type AuthTokenNotFoundError struct {
}

func (e AuthTokenNotFoundError) Error() string {
	return "Auth token not found"
}

type AuthTokenExpiredError struct {
}

func (e AuthTokenExpiredError) Error() string {
	return "Auth token expired"
}

type InternalError struct {
	Message string
}

func (e InternalError) Error() string {
	return e.Message
}

type ProviderAuthError struct {
	Message string
}

func (e ProviderAuthError) Error() string {
	return e.Message
}

type UserAuthError struct {
	Message string
}

func (e UserAuthError) Error() string {
	return e.Message
}
