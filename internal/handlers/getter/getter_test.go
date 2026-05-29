package getter_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"shortener/internal/handlers/getter"
	"shortener/internal/repository"

	"github.com/go-chi/chi/v5"
)

// Мок репозиторий
type mockRepo struct {
	getFunc func(hash string) (string, error)
}

func (m *mockRepo) SaveURL(url, hash string) error {
	return nil
}

func (m *mockRepo) GetURL(hash string) (string, error) {
	if m.getFunc != nil {
		return m.getFunc(hash)
	}
	return "", repository.ErrURLNotFound
}
func (m *mockRepo) GetHashByURL(url string) (string, error) {
	return "", nil
}

func (m *mockRepo) Close() error {
	return nil
}

// Создаем мини-роутер чтобы chi сам распарсил
func doChiRequest(handler http.HandlerFunc, path string) *httptest.ResponseRecorder {
	r := chi.NewRouter()
	r.Get("/{hash}", handler) // Регистрируем хендлер с параметром

	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()


	r.ServeHTTP(w, req)

	return w
}

// GET должен вернуть оригинальный URL по сокращённому
func TestGetHandler_Success(t *testing.T) {
	expectedURL := "https://golang.org"
	repo := &mockRepo{
		getFunc: func(hash string) (string, error) {
			return expectedURL, nil
		},
	}

	handler := getter.New(repo)

	w := doChiRequest(handler, "/Abc123XyZ_")

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, expectedURL) {
		t.Errorf("response = %q, want %q", body, expectedURL)
	}
	if !strings.Contains(body, `"original_url"`) {
		t.Errorf("response = %q, want field 'original_url'", body)
	}
}

// Если хэш не найден то ошибка 404

func TestGetHandler_NotFound(t *testing.T) {

	repo := &mockRepo{
		getFunc: func(hash string) (string, error) {
			return "", repository.ErrURLNotFound
		},
	}

	handler := getter.New(repo)

	w := doChiRequest(handler, "/Abcdefghij")

	if w.Code != http.StatusNotFound {
		t.Errorf("статус = %d, а ожидался 404", w.Code)
	}
}

// Невалидный формат хэша значит 400
func TestGetHandler_InvalidHashFormat(t *testing.T) {
	repo := &mockRepo{}
	handler := getter.New(repo)

	tests := []struct {
		name string
		path string
	}{
		{"too short", "/abc"},         // 3 символа -> 400
		{"too long", "/abcdefghijk"},  // 11 символов -> 400
		{"with dot", "/abc.defghij"},  // 10 символов, но есть точка -> 400
		{"with dash", "/abc-defghij"}, // 10 символов, но есть дефис -> 400
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := doChiRequest(handler, tt.path)

			if w.Code != http.StatusBadRequest {
				t.Errorf("%s: status = %d, want %d", tt.name, w.Code, http.StatusBadRequest)
			}


			if !strings.Contains(w.Body.String(), "invalid hash format") {
				t.Errorf("%s: response = %q, want 'invalid hash format'", tt.name, w.Body.String())
			}
		})
	}
}

// Внутренная ошибка БД 500-тка
func TestGetHandler_InternalError(t *testing.T) {
	repo := &mockRepo{
		getFunc: func(hash string) (string, error) {
			return "", errors.New("database connection lost")
		},
	}

	handler := getter.New(repo)
	w := doChiRequest(handler, "/Abc123XyZ_")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(w.Body.String(), "internal error") {
		t.Errorf("response = %q, want 'internal error'", w.Body.String())
	}
}
