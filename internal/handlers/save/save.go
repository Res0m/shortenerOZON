package save

import (
	"errors"
	"net/http"
	"shortener/generator"
	"shortener/internal/repository"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL string `json:"url" validate:"required,url"`
}

type Response struct {
	Hash string `json:"alias,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func New(repo repository.Repository, size int, validate *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Request

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, ErrorResponse{Error: "invalid request body"})
			return
		}

		if err := validate.Struct(req); err != nil {
			var ve validator.ValidationErrors
			if ok := errors.As(err, &ve); !ok {
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, ErrorResponse{Error: "validation failed"})
				return
			}
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, ErrorResponse{
				Error: "invalid field: " + ve[0].Field(),
			})
			return
		}

		existingHash, err := repo.GetHashByURL(req.URL)
        if err == nil {
            // Ссылка уже есть! Возвращаем старый хеш.
            // Статус 200 OK или 201 Created — на твой выбор. 200 честнее для "найдено".
            render.Status(r, http.StatusOK) 
            render.JSON(w, r, Response{Hash: existingHash})
            return

		}

		var maxRetries int = 5
		var hash string

		for i := 0; i < maxRetries; i++ {
			hash, err = generator.Generate(size)
			if err != nil {
				break
			}
			err = repo.SaveURL(req.URL, hash)
			if err == nil {
				break
			}
			if errors.Is(err, repository.ErrHashExists) {
				continue
			}
			break
		}

		if err != nil {
			switch {
			case errors.Is(err, repository.ErrURLExists):
				render.Status(r, http.StatusConflict)
				render.JSON(w, r, ErrorResponse{Error: "url already exists"})
			case errors.Is(err, repository.ErrHashExists):
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, ErrorResponse{Error: "failed to generate unique alias"})
			default:
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, ErrorResponse{Error: "failed to save url"})
			}
			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{
			Hash: hash,
		})
	}
}
