package auth

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kw3a/spotted-server/internal/database"
)

type AuthStorage struct {
	DB *database.Queries
}

func NewAuthStorage(dbURL string) (*AuthStorage, error) {
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	queries := database.New(db)
	return &AuthStorage{DB: queries}, nil
}

func (a *AuthStorage) GetUserID(ctx context.Context, params AuthStorageParams) (string, error) {
	dbUser, err := a.DB.GetUserByEmail(ctx, params.Email)
	if err != nil {
		return "", errors.New("email and password do not match")
	}
	user := NewUser(dbUser.ID, dbUser.Email, dbUser.Password)
	err = user.CheckPasswordHash(params.Password)
	if err != nil {
		return "", errors.New("email and password do not match")
	}
	return user.ID, nil
}
func (a *AuthStorage) IsRegistered(ctx context.Context, refreshToken string) error {
	tokenExists, err := a.DB.VerifyRefreshJWT(ctx, refreshToken)
	if err != nil {
		return err
	}
	if !tokenExists {
		return errors.New("token does not exist")
	}
	return nil
}
func (a *AuthStorage) Save(ctx context.Context, refreshToken string) error {
	return a.DB.SaveRefreshJWT(ctx, database.SaveRefreshJWTParams{
		RefreshToken: refreshToken,
		CreatedAt:    time.Now().UTC(),
	})
}
func (a *AuthStorage) Revoke(ctx context.Context, refreshToken string) error {
	err := a.IsRegistered(ctx, refreshToken)
	if err != nil {
		return err
	}
	return a.DB.RevokeRefreshJWT(ctx, refreshToken)
}

func (a *AuthStorage) GetRole(ctx context.Context, userID string) (string, error) {
	user, err := a.DB.GetUser(ctx, userID)
	if err != nil {
		return "", err
	}
	return user.Role, nil
}
