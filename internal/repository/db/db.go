package db

import (
	"database/sql"
	"errors"
	"fmt"
	"shortener/internal/config"
	"shortener/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Db struct {
	*sql.DB
}

func NewDb(cfg *config.Config) (*Db, error) {
	dsn := cfg.Dsn.DSN
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("Error to connect DB: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Error of checking connection to DB: %w", err)
	}
	query := `
CREATE TABLE IF NOT EXISTS urlT (
    id SERIAL PRIMARY KEY,
    hash TEXT NOT NULL,
    url TEXT NOT NULL,
    CONSTRAINT urlT_hash_unique UNIQUE (hash),
    CONSTRAINT urlT_url_unique UNIQUE (url)
);`
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("Error to do query: %w", err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("err: %w", err)
	}

	indexQuery := "CREATE INDEX IF NOT EXISTS idx_hash ON urlT(hash);"

	stmt, err = db.Prepare(indexQuery)
	if err != nil {
		return nil, fmt.Errorf("Error to do query: %w", err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("err: %w", err)
	}

	fmt.Println("Succsesfull connection to DB")

	return &Db{DB: db}, nil
}

func (d *Db) SaveURL(urlToSave, hash string) error {

	query := "INSERT INTO urlT(url, hash) VALUES($1, $2) RETURNING id"

	var id int64
	err := d.QueryRow(query, urlToSave, hash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "urlT_url_unique":
				return repository.ErrURLExists
			case "urlT_hash_unique":
				return repository.ErrHashExists
			default:
				return fmt.Errorf("unique constraint violated: %s", pgErr.ConstraintName)
			}
		}
		return fmt.Errorf("failed to insert url: %w", err)
	}
	return nil
}

func (d *Db) GetURL(hash string) (string, error) {
	query := "SELECT url FROM urlT WHERE hash = $1"
	var res string
	err := d.QueryRow(query, hash).Scan(&res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", repository.ErrURLNotFound
		}
		return "", fmt.Errorf("database error: %v", err)
	}
	return res, nil
}


func (d *Db) GetHashByURL(url string) (string, error) {
    var hash string
    err := d.QueryRow("SELECT hash FROM urlT WHERE url = $1", url).Scan(&hash)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return "", repository.ErrURLNotFound
        }
        return "", fmt.Errorf("db error: %w", err)
    }
    return hash, nil
}

func (d *Db) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}
