package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/nuric/go-api-template/models"
	"gorm.io/gorm"
)

type contextKey string

const sessionName = "app-session"
const userKey contextKey = "currentUser"
const userIDKey = "userId"

func UserMiddleware(next http.Handler, db *gorm.DB, store sessions.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, err := store.Get(r, sessionName)
		if err != nil {
			slog.Error("Failed to get session", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// for k, v := range s.Values {
		// 	log.Debug().Any("key", k).Str("kt", fmt.Sprintf("%T", k)).Interface("value", v).Str("t", fmt.Sprintf("%T", v)).Msg("Session value")
		// }
		// Check if user Id is set in the session
		if userId, ok := s.Values[userIDKey].(uint); ok {
			// Fetch user from database to ensure user exists
			var user models.User
			if err := db.First(&user, userId).Error; err != nil && err != gorm.ErrRecordNotFound {
				slog.Error("Failed to fetch user from database", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			// Store user ID in request context for further use
			ctx := r.Context()
			ctx = context.WithValue(ctx, userKey, user)
			r = r.WithContext(ctx)
		}
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func AuthenticatedOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetCurrentUser(r)
		if user.ID == 0 || user.Email == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func VerifiedOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetCurrentUser(r)
		if user.ID == 0 || user.Email == "" || !user.EmailVerified {
			http.Redirect(w, r, "/verify-email", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetCurrentUser(r *http.Request) models.User {
	user, ok := r.Context().Value(userKey).(models.User)
	if !ok {
		return models.User{}
	}
	return user
}

func LogUserIn(w http.ResponseWriter, r *http.Request, userId uint, store sessions.Store) error {
	s, err := store.New(r, sessionName)
	if err != nil {
		slog.Error("Failed to create session", "error", err)
		return err
	}
	s.Values[userIDKey] = userId
	if err := s.Save(r, w); err != nil {
		slog.Error("Failed to save session", "error", err)
		return err
	}
	slog.Debug("User logged in", "userId", userId)
	return nil
}

func LogUserOut(w http.ResponseWriter, r *http.Request, store sessions.Store) error {
	s, err := store.Get(r, sessionName)
	if err != nil {
		slog.Error("Failed to get session", "error", err)
		return err
	}
	s.Values = make(map[any]any) // Clear session values
	if err := s.Save(r, w); err != nil {
		slog.Error("Failed to save session", "error", err)
		return err
	}
	slog.Debug("User logged out")
	return nil
}
