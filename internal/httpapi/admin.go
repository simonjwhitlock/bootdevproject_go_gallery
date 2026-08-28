package httpapi

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/simonjwhitlock/bootdevproject_go_gallery/internal/auth"
	"github.com/simonjwhitlock/bootdevproject_go_gallery/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type AdminHandler struct {
	DB            *database.Queries
	TokenSecret   string
	TokenDuration time.Duration
}

func (h *AdminHandler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.DB.GetUserByEmail(req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		log.Printf("Login error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email, h.TokenSecret, h.TokenDuration)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, user.Email, h.TokenSecret, h.TokenDuration*7)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user_id":       user.ID.String(),
		"email":         user.Email,
	})
}

func (h *AdminHandler) ListImages(w http.ResponseWriter, r *http.Request) {
	images, err := h.DB.ListImages()
	if err != nil {
		log.Printf("ListImages error: %v", err)
		http.Error(w, "Failed to fetch images", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(images)
}

func (h *AdminHandler) CreateImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ImageName        string  `json:"image_name"`
		ImageURL         string  `json:"image_url"`
		ThumbnailURL     string  `json:"thumbnail_url"`
		ImageDescription *string `json:"image_description"`
		DisplayOrder     int     `json:"display_order"`
		UserID           string  `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	now := time.Now()
	img := &database.Image{
		ID:               uuid.New(),
		CreatedAt:        now,
		UpdatedAt:        now,
		ImageName:        req.ImageName,
		ImageURL:         req.ImageURL,
		ThumbnailURL:     req.ThumbnailURL,
		ImageDescription: req.ImageDescription,
		DisplayOrder:     req.DisplayOrder,
		UserID:           userID,
	}

	created, err := h.DB.CreateImage(img)
	if err != nil {
		log.Printf("CreateImage error: %v", err)
		http.Error(w, "Failed to create image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *AdminHandler) UpdateImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID               string  `json:"id"`
		ImageName        string  `json:"image_name"`
		ImageURL         string  `json:"image_url"`
		ThumbnailURL     string  `json:"thumbnail_url"`
		ImageDescription *string `json:"image_description"`
		DisplayOrder     int     `json:"display_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		http.Error(w, "Invalid image id", http.StatusBadRequest)
		return
	}

	img, err := h.DB.GetImage(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}
		log.Printf("GetImage error: %v", err)
		http.Error(w, "Failed to fetch image", http.StatusInternalServerError)
		return
	}

	img.ImageName = req.ImageName
	img.ImageURL = req.ImageURL
	img.ThumbnailURL = req.ThumbnailURL
	img.ImageDescription = req.ImageDescription
	img.DisplayOrder = req.DisplayOrder
	img.UpdatedAt = time.Now()

	updated, err := h.DB.UpdateImage(img)
	if err != nil {
		log.Printf("UpdateImage error: %v", err)
		http.Error(w, "Failed to update image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *AdminHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	if err := h.DB.DeleteImage(id); err != nil {
		log.Printf("DeleteImage error: %v", err)
		http.Error(w, "Failed to delete image", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) ReorderImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Images []struct {
			ID           string `json:"id"`
			DisplayOrder int    `json:"display_order"`
		} `json:"images"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	now := time.Now()
	for _, img := range req.Images {
		id, err := uuid.Parse(img.ID)
		if err != nil {
			continue
		}
		if err := h.DB.ReorderImages(id, img.DisplayOrder, now); err != nil {
			log.Printf("ReorderImages error for %s: %v", img.ID, err)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
