package auth

import (
	"context"
	"fmt"
	"net/http"
)

type AuthUser struct {
	ID   string
	Role string
}

func (auth *AuthService) GetUser(r *http.Request) (userID string, err error) {
	ctx := r.Context()
	user, ok := ctx.Value(AuthUser{}).(AuthUser)
	if !ok {
		return "", fmt.Errorf("no user in context")
	}
	role, err := auth.Storage.GetRole(ctx, user.ID)
	if err != nil {
		return "", err
	}
	if role != user.Role {
		return "", fmt.Errorf("role mismatch")
	}
	return user.ID, nil
}

func (auth *AuthService) Middleware(role string, next http.HandlerFunc) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("access_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		access_token := token.Value
		userID, err := auth.JWTRep.Authenticate(access_token)
		if err != nil {
      refresh, err := r.Cookie("refresh_token")
      if err != nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
      }
			access_token, err = auth.JWTRep.CreateAccess(refresh.Value)
      if err != nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
      }
      http.SetCookie(w, &http.Cookie{
        Name:    "access,token",
        Value:   access_token,
        Secure:  true,
        HttpOnly: true,
        SameSite: http.SameSiteLaxMode,
      })
		}
		authUser := AuthUser{
			ID:   userID,
			Role: role,
		}
		ctx := context.WithValue(r.Context(), AuthUser{}, authUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (auth *AuthService) DevMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return auth.Middleware("dev", next)
}
