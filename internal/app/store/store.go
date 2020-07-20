package store

import (
	"database/sql"

	_ "github.com/lib/pq" //  ...
)

// Store ...
type Store struct {
	config *Config
	db     *sql.DB
	User   *UserRepository
	News   *NewsRepository
}

// New ...
func New(config *Config) *Store {
	store := &Store{
		config: config,
	}
	store.User = &UserRepository{store: store}
	store.News = &NewsRepository{store: store}
	return store
}

// Open ...
func (store *Store) Open() error {
	db, err := sql.Open("postgres", store.config.DSN)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}
	store.db = db
	return nil
}

// Close ...
func (store *Store) Close() {
	store.db.Close()
}
