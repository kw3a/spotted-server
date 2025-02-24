package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kw3a/spotted-server/internal/database"
)

type AuthUser struct {
	ID       string
	Role     string
	Name     string
	ImageURL string
	Email    string
}
type AuthService struct{}

func (a *AuthService) GetUser(r *http.Request) (AuthUser, error) {
	ctx := r.Context()
	user, ok := ctx.Value(AuthUser{}).(AuthUser)
	if !ok {
		return AuthUser{}, fmt.Errorf("no user in context")
	}
	return user, nil
}

func getCookies(r *http.Request) (string, string, error) {
	access, err := r.Cookie("access_token")
	if err != nil {
		return "", "", err
	}
	refresh, err := r.Cookie("refresh_token")
	if err != nil {
		return "", "", err
	}
	return refresh.Value, access.Value, nil
}
func SetTokenCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
func deleteTokenCookie(w http.ResponseWriter, name string) {
	expired := time.Now().Add(-time.Hour)
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Expires:  expired,
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
func DeleteCookies(w http.ResponseWriter) {
	deleteTokenCookie(w, "refresh_token")
	deleteTokenCookie(w, "access_token")
}

type MiddlewareStorage interface {
	GetUser(ctx context.Context, userID string) (database.User, error)
}
type MiddlewareAuthType interface {
	WhoAmI(accessToken string) (userID string, err error)
	CreateAccess(refreshToken string) (string, error)
}

func CreateMiddleware(storage MiddlewareStorage, authType MiddlewareAuthType, loginPath, role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refresh_token, access_token, err := getCookies(r)
		if err != nil {
			http.Redirect(w, r, loginPath, http.StatusSeeOther)
			return
		}
		userID, err := authType.WhoAmI(access_token)
		if err != nil {
			access_token, err = authType.CreateAccess(refresh_token)
			if err != nil {
				http.Redirect(w, r, loginPath, http.StatusSeeOther)
				return
			}
			SetTokenCookie(w, "access_token", access_token)
			userID, err = authType.WhoAmI(access_token)
			if err != nil {
				http.Redirect(w, r, loginPath, http.StatusSeeOther)
				return
			}
		}
		dbUser, err := storage.GetUser(r.Context(), userID)
		if err != nil || role != dbUser.Role {
			DeleteCookies(w)
			http.Redirect(w, r, loginPath, http.StatusSeeOther)
			return
		}
		ctx := context.WithValue(r.Context(), AuthUser{}, AuthUser{
			ID:   userID,
			Role: role,
			Name: dbUser.Name,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AuthNMiddleware(storage MiddlewareStorage, authType MiddlewareAuthType, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setContext := func(user AuthUser) {
			ctx := context.WithValue(r.Context(), AuthUser{}, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		refresh_token, access_token, err := getCookies(r)
		if err != nil {
			setContext(AuthUser{Role: "visitor"})
			return
		}
		userID, err := authType.WhoAmI(access_token)
		if err != nil {
			access_token, err = authType.CreateAccess(refresh_token)
			if err != nil {
				setContext(AuthUser{Role: "visitor"})
				return
			}
			SetTokenCookie(w, "access_token", access_token)
			userID, err = authType.WhoAmI(access_token)
			if err != nil {
				setContext(AuthUser{Role: "visitor"})
				return
			}
		}
		dbUser, err := storage.GetUser(r.Context(), userID)
		if err != nil {
			setContext(AuthUser{Role: "visitor"})
			return
		}
		setContext(AuthUser{
			ID:       dbUser.ID,
			Role:     dbUser.Role,
			Name:     dbUser.Name,
			ImageURL: dbUser.ImageUrl,
			Email:    dbUser.Email,
		})
	})
}

func AuthRMiddleware(loginPath, role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := ctx.Value(AuthUser{}).(AuthUser)
		if !ok {
			http.Redirect(w, r, loginPath, http.StatusSeeOther)
			return
		}
		if role != user.Role {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
