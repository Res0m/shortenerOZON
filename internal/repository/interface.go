package repository

import "errors"

var (
	ErrURLNotFound = errors.New("url not found")
	ErrURLExists   = errors.New("url already exists")
	ErrHashExists  = errors.New("hash already exists")
)

type Repository interface {
	SaveURL(url, hash string) error
	GetURL(hash string) (string, error)
	GetHashByURL(url string) (string, error)
	Close() error
}