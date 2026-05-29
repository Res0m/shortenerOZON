package memory_test

import (
	"errors"
	"sync"
	"testing"

	"shortener/internal/repository"
	"shortener/internal/repository/memory"
)

// Проверка что один URL один hash
func TestMemory_SaveURL_SameURLSameHash(t *testing.T) {
	m := memory.New()
	url := "https://example.com"
	hash := "Abc123XyZ_"

	// 1)
	if err := m.SaveURL(url, hash); err != nil {
		t.Fatalf("SaveURL() first = %v, want nil", err)
	}

	// 2)
	if err := m.SaveURL(url, hash); err != nil {
		t.Errorf("SaveURL() duplicate = %v, want nil", err)
	}

	// Проверяем, что можно получить оригинал по хешу
	got, err := m.GetURL(hash)
	if err != nil {
		t.Fatalf("GetURL() error: %v", err)
	}
	if got != url {
		t.Errorf("GetURL() = %q, want %q", got, url)
	}
}

// Проверка, что если уже есть такой URL с  таким хэшом, то нельзя сохранить с другим
func TestMemory_SaveURL_DifferentHash(t *testing.T) {
	m := memory.New()
	url := "https://example.com"

	// Сохраняем с первым хешом
	_ = m.SaveURL(url, "FirstHash123")

	// Пытаемся сохранить тот же URL с другим хешом это ошибка
	err := m.SaveURL(url, "OtherHash456")
	if !errors.Is(err, repository.ErrURLExists) {
		t.Errorf("SaveURL() error = %v, want ErrURLExists", err)
	}
}

// Если хэш занят другим URL то это ошибка
func TestMemory_SaveURL_DifferentURL(t *testing.T) {
	m := memory.New()
	hash := "SameHash1234"

	// Сохраняем первый URL с этим хешом
	_ = m.SaveURL("https://first.com", hash)

	// Пытаемся сохранить другой URL с тем же хешом это ошибка
	err := m.SaveURL("https://second.com", hash)
	if !errors.Is(err, repository.ErrHashExists) {
		t.Errorf("SaveURL() error = %v, want ErrHashExists", err)
	}
}

// Должен вернуть оригинал по хэшу
func TestMemory_GetURL_Success(t *testing.T) {
	m := memory.New()
	url := "https://golang.org"
	hash := "GoLang1234"

	_ = m.SaveURL(url, hash)

	got, err := m.GetURL(hash)
	if err != nil {
		t.Fatalf("GetURL() error: %v", err)
	}
	if got != url {
		t.Errorf("GetURL() = %q, want %q", got, url)
	}
}

// Если хэша не найдено то это ошибка
func TestMemory_GetURL_NotFound(t *testing.T) {
	m := memory.New()

	_, err := m.GetURL("NotExist123")
	if !errors.Is(err, repository.ErrURLNotFound) {
		t.Errorf("GetURL() error = %v, want ErrURLNotFound", err)
	}
}

// Хранилище должно корректно работать при конкурентном доступе
func TestMemory_Concurrent(t *testing.T) {
	m := memory.New()
	const goroutines = 50
	const opsPerGoroutine = 20

	var wg sync.WaitGroup

	// Параллельно пишем и читаем
	for i := 0; i < goroutines; i++ {
		wg.Add(2)

		// Горутина на запись
		go func(id int) {
			defer wg.Done()
			url := "https://example.com/" + string(rune(id))
			hash := "Hash" + string(rune(id)) + "012345678" 
			_ = m.SaveURL(url, hash)                        
		}(i)

		// Горутина на чтение
		go func(id int) {
			defer wg.Done()
			hash := "Hash" + string(rune(id)) + "012345678"
			_, _ = m.GetURL(hash) 
		}(i)
	}

	wg.Wait()
}


// Тест функции GetHashByUrl
func TestMemory_GetHashByURL_Success(t *testing.T) {
	m := memory.New()
	url := "https://example.com"
	hash := "TestHash1234" // 10 символов
	

	if err := m.SaveURL(url, hash); err != nil {
		t.Fatalf("SaveURL() error: %v", err)
	}
	

	gotHash, err := m.GetHashByURL(url)
	if err != nil {
		t.Fatalf("GetHashByURL() error: %v", err)
	}
	
	if gotHash != hash {
		t.Errorf("GetHashByURL() = %q, want %q", gotHash, hash)
	}
}

// Тест на несуществующие URL 
func TestMemory_GetHashByURL_NotFound(t *testing.T) {
	m := memory.New()
	

	_, err := m.GetHashByURL("https://non-existent.com")
	
	if !errors.Is(err, repository.ErrURLNotFound) {
		t.Errorf("GetHashByURL() error = %v, want ErrURLNotFound", err)
	}
}


// Проверка Закрытия
func TestMemory_Close(t *testing.T) {
	m := memory.New()
	if err := m.Close(); err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}
