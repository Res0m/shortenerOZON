package memory

import (
	"shortener/internal/repository"
	"sync"
)

type Memory struct {
	mu        sync.RWMutex
	hashToURL map[string]string
	urlToHash map[string]string 
}

func New() *Memory {
	return &Memory{
		hashToURL: make(map[string]string),
		urlToHash: make(map[string]string),
	}
}


func (m *Memory) SaveURL(url, hash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()


	if existingHash, ok := m.urlToHash[url]; ok {
		if existingHash == hash {
			return nil 
		}

		return repository.ErrURLExists
	}


	if existingURL, ok := m.hashToURL[hash]; ok {
		if existingURL == url {
			return nil 
		}
		
		return repository.ErrHashExists
	}

	
	m.hashToURL[hash] = url
	m.urlToHash[url] = hash

	return nil
}


func (m *Memory) GetURL(hash string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	url, ok := m.hashToURL[hash]
	if !ok {
		return "", repository.ErrURLNotFound
	}

	return url, nil
}


func (m *Memory) GetHashByURL(url string) (string, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    hash, ok := m.urlToHash[url]
    if !ok {
        return "", repository.ErrURLNotFound
    }
    return hash, nil
}


func (m *Memory) Close() error {
	return nil
}
