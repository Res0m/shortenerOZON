package save_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"shortener/internal/handlers/save"
	"shortener/internal/repository"

	"github.com/go-playground/validator/v10"
)

// Мок репозиторий
type mockRepo struct {
	saveFunc func(url, hash string) error
}

func (m *mockRepo) SaveURL(url, hash string) error {
	if m.saveFunc != nil {
		return m.saveFunc(url, hash)
	}
	return nil
}

func (m *mockRepo) GetURL(hash string) (string, error) {
	return "", nil
}

func (m *mockRepo) GetHashByURL(url string) (string, error) {
	return "", nil
}

func (m *mockRepo) Close() error {
	return nil
}

// Функция для выставления запроса
func doRequest(handler http.HandlerFunc, method, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	return w
}

// Обработка успешного сценария POST запроса
func TestSaveHandler_Success(t *testing.T) {
	repo := &mockRepo{
		saveFunc: func(url, hash string) error {
			// Проверяем, что хеш имеет нужную длину и символы
			if len(hash) != 10 {
				t.Errorf("hash length = %d, want 10", len(hash))
			}
			for _, c := range hash {
				if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
					t.Errorf("hash contains invalid character: %c", c)
				}
			}
			return nil
		},
	}
	validate := validator.New()
	handler := save.New(repo, 10, validate)

	body := `{"url": "https://golang.org"}`
	w := doRequest(handler, http.MethodPost, body)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
	if !strings.Contains(w.Body.String(), `"alias"`) {
		t.Errorf("response = %q, want field 'alias'", w.Body.String())
	}
}

// Обработка невалидного JSON
func TestSaveHandler_InvalidJSON(t *testing.T) {
	repo := &mockRepo{}
	validate := validator.New()
	handler := save.New(repo, 10, validate)

	body := `{"url": "invalid json` // типо незакрытая ковычка
	w := doRequest(handler, http.MethodPost, body)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if !strings.Contains(w.Body.String(), "invalid request body") {
		t.Errorf("response = %q, want error message", w.Body.String())
	}
}

// Обработка невалидного URL
func TestSaveHandler_InvalidURL(t *testing.T) {
	repo := &mockRepo{}
	validate := validator.New()
	handler := save.New(repo, 10, validate)

	body := `{"url": "not-a-url"}`
	w := doRequest(handler, http.MethodPost, body)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if !strings.Contains(w.Body.String(), "invalid field") {
		t.Errorf("response = %q, want validation error", w.Body.String())
	}
}

// Обработка пустого URL
func TestSaveHandler_EmptyURL(t *testing.T) {
	repo := &mockRepo{}
	validate := validator.New()
	handler := save.New(repo, 10, validate)

	body := `{"url": ""}`
	w := doRequest(handler, http.MethodPost, body)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// один оригинальный URL - одна сокращенная ссылка
func TestSaveHandler_URLOnlyExists(t *testing.T) {
	repo := &mockRepo{
		saveFunc: func(url, hash string) error {
			return repository.ErrURLExists
		},
	}
	validate := validator.New()
	handler := save.New(repo, 10, validate)

	body := `{"url": "https://already-exists.com"}`
	w := doRequest(handler, http.MethodPost, body)

	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", w.Code, http.StatusConflict)
	}
	if !strings.Contains(w.Body.String(), "url already exists") {
		t.Errorf("response = %q, want conflict message", w.Body.String())
	}
}

// Обработка коллизий хэша
func TestSaveHandler_HashCollision_Retry(t *testing.T) {
	attempts := 0
	repo := &mockRepo{
		saveFunc: func(url, hash string) error {
			attempts++
			if attempts < 3 {
				return repository.ErrHashExists
			}
			return nil
		},
	}
	validate := validator.New()
	handler := save.New(repo, 10, validate)

	body := `{"url": "https://retry-test.com"}`
	w := doRequest(handler, http.MethodPost, body)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3 (2 collisions + 1 success)", attempts)
	}
}

// Если много коллизий это плохо
func TestSaveHandler_TooManyCollisions(t *testing.T) {
	repo := &mockRepo{
		saveFunc: func(url, hash string) error {
			return repository.ErrHashExists
		},
	}
	validate := validator.New()
	handler := save.New(repo, 10, validate)

	body := `{"url": "https://always-collide.com"}`
	w := doRequest(handler, http.MethodPost, body)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(w.Body.String(), "failed to generate unique alias") {
		t.Errorf("response = %q, want collision error", w.Body.String())
	}
}

// Обработка ошибки генератора
func TestSaveHandler_GeneratorError(t *testing.T) {
	repo := &mockRepo{
		saveFunc: func(url, hash string) error {
			return errors.New("unexpected db error")
		},
	}
	validate := validator.New()
	handler := save.New(repo, 10, validate)

	body := `{"url": "https://example.com"}`
	w := doRequest(handler, http.MethodPost, body)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}



