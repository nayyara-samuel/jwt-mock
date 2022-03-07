package handlers

import (
	"net/http"

	"github.com/nayyara-cropsey/jwt-mock/jwt"
	"github.com/nayyara-cropsey/jwt-mock/service"

	"go.uber.org/zap"
)

type jwtResponse struct {
	Token string `json:"token"`
}

// JWTDefaultPath is the default path for JWT handlers.
const JWTDefaultPath = "/generate-jwt"

// JWTHandler provides handlers for working with JWTs
type JWTHandler struct {
	keyStore *service.KeyStore
	logger   *zap.Logger
}

// NewJWTHandler is the preferred way to create a JWTHandler instance.
func NewJWTHandler(keyStore *service.KeyStore, logger *zap.Logger) *JWTHandler {
	return &JWTHandler{
		keyStore: keyStore,
		logger:   logger,
	}
}

// RegisterDefaultPaths registers the default paths for JWKS operations.
func (j *JWTHandler) RegisterDefaultPaths(api *http.ServeMux) {
	api.HandleFunc(JWTDefaultPath, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			j.Post(w, r)
		default:
			notFoundResponse(w)
		}
	})
}

// Post creates a signed JWT with the provided claims.
func (j *JWTHandler) Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var claims jwt.Claims
	if err := jsonUnmarshal(r, &claims); err != nil {
		j.logger.Error("failed to read claims", zap.Error(err))

		w.WriteHeader(http.StatusBadRequest)

		if err = jsonMarshal(w, errorResponse{
			Message: "Failed to read claims",
			Error:   err.Error(),
		}); err != nil {
			j.logger.Error("Failed write JSON response", zap.Error(err))
		}

		return
	}

	signingKey := j.keyStore.GetSigningKey()
	token, err := jwt.CreateToken(claims, signingKey)
	if err != nil {
		j.logger.Error("failed to generate JWT", zap.Error(err))

		w.WriteHeader(http.StatusBadRequest)

		if err = jsonMarshal(w, errorResponse{
			Message: "Failed to generate JWT",
			Error:   err.Error(),
		}); err != nil {
			j.logger.Error("Failed write JSON response", zap.Error(err))
		}

		return
	}

	if err := jsonMarshal(w, jwtResponse{Token: token}); err != nil {
		j.logger.Error("Failed write JSON response", zap.Error(err))
		return
	}

	w.WriteHeader(http.StatusOK)
}
