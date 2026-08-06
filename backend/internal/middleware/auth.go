package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/rotaract-civ/backend/internal/auth"
	"github.com/rotaract-civ/backend/internal/authctx"
	"github.com/rotaract-civ/backend/internal/domain"
	"github.com/rotaract-civ/backend/internal/handler"
	"github.com/rotaract-civ/backend/internal/repository"
)

func Authenticate(tokens *auth.TokenManager, users *repository.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, errMsg, err := resolveBearerUser(r, tokens, users)
			if err != nil {
				handler.WriteError(w, http.StatusUnauthorized, errMsg)
				return
			}
			if !user.IsActive {
				handler.WriteError(w, http.StatusForbidden, "account is inactive")
				return
			}
			ctx := authctx.WithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthenticate attaches the user when a valid Bearer token is present.
// Guests continue without a user in context.
func OptionalAuthenticate(tokens *auth.TokenManager, users *repository.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, _, err := resolveBearerUser(r, tokens, users)
			if err == nil && user != nil && user.IsActive {
				r = r.WithContext(authctx.WithUser(r.Context(), user))
			}
			next.ServeHTTP(w, r)
		})
	}
}

func resolveBearerUser(r *http.Request, tokens *auth.TokenManager, users *repository.UserRepository) (*domain.User, string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return nil, "missing authorization header", errors.New("missing")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, "invalid authorization header", errors.New("invalid header")
	}

	claims, err := tokens.Parse(parts[1])
	if err != nil {
		return nil, "invalid or expired token", err
	}

	user, err := users.GetByID(r.Context(), claims.UserID)
	if err != nil {
		return nil, "user not found", err
	}
	return user, "", nil
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := authctx.UserFromContext(r.Context())
		if !ok || !user.IsAdmin {
			handler.WriteError(w, http.StatusForbidden, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireClubPermission(clubs *repository.ClubRepository, permissionKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := authctx.UserFromContext(r.Context())
			if !ok {
				handler.WriteError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			if user.IsAdmin {
				next.ServeHTTP(w, r)
				return
			}

			clubID, err := handler.ClubIDFromRequest(r)
			if err != nil {
				handler.WriteError(w, http.StatusBadRequest, err.Error())
				return
			}

			allowed, err := clubs.UserHasPermission(r.Context(), clubID, user.ID, permissionKey)
			if err != nil {
				handler.WriteError(w, http.StatusInternalServerError, "failed to verify permissions")
				return
			}
			if !allowed {
				handler.WriteError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireClubHead(clubs *repository.ClubRepository) func(http.Handler) http.Handler {
	return RequireClubPermission(clubs, "club.members.create")
}

func RequireCommissionPresident(commissions *repository.CommissionRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := authctx.UserFromContext(r.Context())
			if !ok {
				handler.WriteError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			if user.IsAdmin {
				next.ServeHTTP(w, r)
				return
			}

			commissionID, err := handler.CommissionIDFromRequest(r)
			if err != nil {
				handler.WriteError(w, http.StatusBadRequest, err.Error())
				return
			}

			isPresident, err := commissions.IsPresident(r.Context(), commissionID, user.ID)
			if err != nil {
				handler.WriteError(w, http.StatusInternalServerError, "failed to verify commission role")
				return
			}
			if !isPresident {
				handler.WriteError(w, http.StatusForbidden, "commission president access required")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ClubContextKey stores resolved club ID for downstream handlers if needed.
type ClubContextKey struct{}

func WithClubID(ctx context.Context, clubID string) context.Context {
	return context.WithValue(ctx, ClubContextKey{}, clubID)
}
