package httpapi

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/simonjwhitlock/bootdevproject_go_gallery/internal/database"
	"github.com/simonjwhitlock/bootdevproject_go_gallery/internal/storage"
)

type PublicHandler struct {
	DB      *database.Queries
	Storage *storage.R2Client
}

func (h *PublicHandler) ListImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	images, err := h.DB.ListImages()
	if err != nil {
		log.Printf("ListImages error: %v", err)
		http.Error(w, "Failed to fetch images", http.StatusInternalServerError)
		return
	}

	// Transform images to include full URLs using storage public URL
	type ImageResponse struct {
		ID               string  `json:"id"`
		CreatedAt        string  `json:"created_at"`
		UpdatedAt        string  `json:"updated_at"`
		ImageName        string  `json:"image_name"`
		ImageURL         string  `json:"image_url"`
		ThumbnailURL     string  `json:"thumbnail_url"`
		ImageDescription *string `json:"image_description"`
		DisplayOrder     int     `json:"display_order"`
		UserID           string  `json:"user_id"`
	}

	responses := make([]ImageResponse, len(images))
	for i, img := range images {
		responses[i] = ImageResponse{
			ID:               img.ID.String(),
			CreatedAt:        img.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:        img.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			ImageName:        img.ImageName,
			ImageURL:         h.Storage.GetPublicURL(img.ImageURL),
			ThumbnailURL:     h.Storage.GetPublicURL(img.ThumbnailURL),
			ImageDescription: img.ImageDescription,
			DisplayOrder:     img.DisplayOrder,
			UserID:           img.UserID.String(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

func (h *PublicHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	images, err := h.DB.ListImages()
	if err != nil {
		log.Printf("ListImages error: %v", err)
		http.Error(w, "Failed to fetch images", http.StatusInternalServerError)
		return
	}

	// Find the image by ID and return with full URL
	for _, img := range images {
		if img.ID.String() == idStr {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":                img.ID.String(),
				"created_at":        img.CreatedAt.Format("2006-01-02T15:04:05Z"),
				"updated_at":        img.UpdatedAt.Format("2006-01-02T15:04:05Z"),
				"image_name":        img.ImageName,
				"image_url":         h.Storage.GetPublicURL(img.ImageURL),
				"thumbnail_url":     h.Storage.GetPublicURL(img.ThumbnailURL),
				"image_description": img.ImageDescription,
				"display_order":     img.DisplayOrder,
				"user_id":           img.UserID.String(),
			})
			return
		}
	}

	http.Error(w, "Image not found", http.StatusNotFound)
}
