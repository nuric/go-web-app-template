package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

// Validator is an object that can be validated.
type Validator interface {
	Validate() error
}

// Encode writes the object to the response writer. It is usually used as the
// last step in a handler.
func Encode[T any](w http.ResponseWriter, status int, v T) {
	w.Header().Set("Content-Type", "application/json")
	// Write to buffer first to ensure the object is json encodable
	// before writing to the response writer.
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Err(err).Msg("could not encode response")
		if errErr := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); errErr != nil {
			log.Error().Err(errErr).Msg("could not encode error response")
			http.Error(w, "could not encode error response", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(status)
	if size, err := w.Write(buf.Bytes()); size != len(buf.Bytes()) || err != nil {
		log.Error().Err(err).Msg("could not write response")
		http.Error(w, "could not write response", http.StatusInternalServerError)
		return
	}
}

// DecodeValid decodes the request body into the object and then validates it.
func DecodeValid[T Validator](r *http.Request) (T, error) {
	var v T
	ctype := r.Header.Get("Content-Type")
	switch ctype {
	case "application/json":
		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			return v, fmt.Errorf("decode json: %w", err)
		}
	default:
		return v, fmt.Errorf("invalid content type %s expected application/json", ctype)
	}
	// ---------------------------
	if err := v.Validate(); err != nil {
		return v, fmt.Errorf("validation error: %w", err)
	}
	// ---------------------------
	return v, nil
}
