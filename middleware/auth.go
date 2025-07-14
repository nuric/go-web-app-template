package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"
)

type contextKey string

const UserIDKey contextKey = "userId"

func AuthenticatedOnly(next http.Handler, store sessions.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, err := store.Get(r, "app-session")
		if err != nil {
			log.Error().Err(err).Msg("Failed to get session")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// for k, v := range s.Values {
		// 	log.Debug().Any("key", k).Str("kt", fmt.Sprintf("%T", k)).Interface("value", v).Str("t", fmt.Sprintf("%T", v)).Msg("Session value")
		// }
		// Check if user Id is set in the session
		userId, ok := s.Values["userId"].(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// Store user ID in request context for further use
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, userId)
		r = r.WithContext(ctx)
		log.Debug().Uint("userId", userId).Msg("Authenticated user")
		// Call the next handler
		next.ServeHTTP(w, r)

	})
}
