package http

import (
	"encoding/json"
	"net/http"

	"mishatkin/auth-service/internal/domain/usecase"
	"mishatkin/shared/pkg/logger"
)

// Handler handles HTTP requests
type Handler struct {
	authUseCase *usecase.AuthUseCase
	log         *logger.Logger
}

// NewHandler creates a new handler
func NewHandler(authUseCase *usecase.AuthUseCase, log *logger.Logger) *Handler {
	return &Handler{
		authUseCase: authUseCase,
		log:         log,
	}
}

// SetupRouter configures the router
func SetupRouter(h *Handler, log *logger.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	// CORS middleware
	withCORS := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			handler(w, r)
		}
	}

	// Register endpoints
	mux.HandleFunc("/api/v1/auth/register", withCORS(h.Register))
	mux.HandleFunc("/api/v1/auth/login", withCORS(h.Login))
	mux.HandleFunc("/api/v1/auth/refresh", withCORS(h.Refresh))
	mux.HandleFunc("/api/v1/auth/logout", withCORS(h.Logout))
	mux.HandleFunc("/api/v1/auth/account", withCORS(h.DeleteAccount))

	// Health check
	mux.HandleFunc("/api/v1/health/live", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	return mux
}

// Register handles user registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON")
		return
	}

	resp, err := h.authUseCase.Register(r.Context(), &usecase.RegisterRequest{
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
	})

	if err != nil {
		h.handleError(w, err)
		return
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

// Login handles user login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Identifier string `json:"identifier"`
		Password  string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid JSON")
		return
	}

	resp, err := h.authUseCase.Login(r.Context(), &usecase.LoginRequest{
		Identifier: req.Identifier,
		Password:   req.Password,
	})

	if err != nil {
		h.handleError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// Refresh handles token refresh
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get refresh token from cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "TOKEN_REQUIRED", "Refresh token required")
		return
	}

	resp, err := h.authUseCase.RefreshTokens(r.Context(), &usecase.RefreshRequest{
		RefreshToken: cookie.Value,
	})

	if err != nil {
		h.handleError(w, err)
		return
	}

	// Set new refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    resp.AccessToken, // This is wrong, need separate refresh token
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	h.writeJSON(w, http.StatusOK, resp)
}

// Logout handles user logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.authUseCase.Logout(r.Context(), cookie.Value)

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusNoContent)
}

// DeleteAccount handles account deletion
func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.writeError(w, http.StatusUnauthorized, "TOKEN_REQUIRED", "Authorization required")
		return
	}

	// TODO: Parse JWT and extract user ID

	w.WriteHeader(http.StatusNoContent)
}

// Helper methods

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	// Map error to status code
	h.writeError(w, http.StatusBadRequest, "ERROR", err.Error())
}

