package auth

import "context"

type AuthService struct {
	Storage AuthenticationStorage
	JWTRep  Authentication
}
type Authentication interface {
	Authenticate(accessToken string) (userID string, err error)
	CreateAccess(refreshToken string) (string, error)
	CreateRefresh(userID string) (string, error)
	ValidateRefresh(refreshToken string) error
}

type AuthStorageParams struct {
	Email    string
	Password string
}
type AuthenticationStorage interface {
	GetUserID(ctx context.Context, params AuthStorageParams) (string, error)
	IsRegistered(ctx context.Context, refreshToken string) error
	Save(ctx context.Context, refreshToken string) error
	Revoke(ctx context.Context, refreshToken string) error
	GetRole(context.Context, string) (string, error)
}

func NewAuthService(secret string, dbURL string) (*AuthService, error) {
	authStorage, err := NewAuthStorage(dbURL)
	if err != nil {
		return nil, err
	}
	authRep := NewJWTAuth(secret)
	return &AuthService{
		Storage: authStorage,
		JWTRep:  authRep,
	}, nil
}
