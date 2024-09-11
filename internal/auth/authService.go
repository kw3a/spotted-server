package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type AuthUser struct {
	ID   string
	Role string
}
type AuthService struct {}
func (a *AuthService) GetUser(r *http.Request) (string, error) {
	ctx := r.Context()
	user, ok := ctx.Value(AuthUser{}).(AuthUser)
	if !ok {
		return "", fmt.Errorf("no user in context")
	}
	return user.ID, nil
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
	GetRole(ctx context.Context, userID string) (string, error)
}
type MiddlewareAuthType interface {
	WhoAmI(accessToken string) (userID string, err error)
	CreateAccess(refreshToken string) (string, error)
}

func CreateMiddleware(storage MiddlewareStorage, authType MiddlewareAuthType, loginPath, role string, next http.Handler) http.Handler{
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
		} else {
			dbRole, err := storage.GetRole(r.Context(), userID)
			if err != nil || role != dbRole {
				DeleteCookies(w)
				http.Redirect(w, r, loginPath, http.StatusSeeOther)
				return
			}
		}
		ctx := context.WithValue(r.Context(), AuthUser{}, AuthUser{
			ID:   userID,
			Role: role,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

