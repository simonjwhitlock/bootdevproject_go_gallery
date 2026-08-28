package httpapi
package httpapi

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/simonjwhitlock/bootdevproject_go_gallery/internal/database"
)

type PublicHandler struct {
	DB *database.Queries
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(images)
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

	// This would need GetImageByID - for now return all and filter on frontend
	images, err := h.DB.ListImages()
	if err != nil {
		log.Printf("ListImages error: %v", err)
		http.Error(w, "Failed to fetch images", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(images)
}
