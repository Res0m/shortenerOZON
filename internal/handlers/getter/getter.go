package getter

import (
	"errors"
	"net/http"
	"regexp"
	"shortener/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	OriginalURL string `json:"original_url"`
}

var hashRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{10}$`)

func New(repo repository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		hash := chi.URLParam(r, "hash")
		if !hashRegex.MatchString(hash) {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, ErrorResponse{Error: "invalid hash format"})
			return
		}

		resURL, err := repo.GetURL(hash)
		if errors.Is(err, repository.ErrURLNotFound) {
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, ErrorResponse{Error: "not found"})
			return
		}

		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, ErrorResponse{Error: "internal error"})
			return
		}

		render.JSON(w, r, SuccessResponse{OriginalURL: resURL})
	}
}
