package storage

import (
	"context"
	"errors"
	"time"

	"github.com/kw3a/spotted-server/internal/database"
	"github.com/kw3a/spotted-server/internal/server/shared"
	"golang.org/x/crypto/bcrypt"
)

func CheckPasswordHash(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
func (s *MysqlStorage) GetUserID(ctx context.Context, nick, password string) (string, error) {
	dbUser, err := s.Queries.GetUserByNick(ctx, nick)
	if err != nil {
		return "", errors.New(shared.ErrNotRegistered)
	}
	err = CheckPasswordHash(dbUser.Password, password)
	if err != nil {
		return "", errors.New(shared.ErrNotMatch)
	}
	return dbUser.ID, nil
}
func (s *MysqlStorage) IsRegistered(ctx context.Context, refreshToken string) error {
	tokenExists, err := s.Queries.VerifyRefreshJWT(ctx, refreshToken)
	if err != nil {
		return err
	}
	if !tokenExists {
		return errors.New("token does not exist")
	}
	return nil
}
func (s *MysqlStorage) Save(ctx context.Context, refreshToken string) error {
	return s.Queries.SaveRefreshJWT(ctx, database.SaveRefreshJWTParams{
		RefreshToken: refreshToken,
		CreatedAt:    time.Now().UTC(),
	})
}
func (s *MysqlStorage) Revoke(ctx context.Context, refreshToken string) error {
	err := s.IsRegistered(ctx, refreshToken)
	if err != nil {
		return err
	}
	return s.Queries.RevokeRefreshJWT(ctx, refreshToken)
}

func (s *MysqlStorage) GetUser(ctx context.Context, userID string) (database.User, error) {
	return s.Queries.GetUser(ctx, userID)
}
