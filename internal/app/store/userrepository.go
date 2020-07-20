package store

import "github.com/NewGlad/xsolla-be/internal/app/model"

// UserRepository ...
type UserRepository struct {
	store *Store
}

// Create ...
func (repository *UserRepository) Create(user *model.User) error {
	if err := user.Validate(); err != nil {
		return err
	}
	if err := user.EncryptPassword(); err != nil {
		return err
	}
	if err := repository.store.db.QueryRow(
		"INSERT INTO users(username, encrypted_password) VALUES ($1, $2) RETURNING id",
		user.Username, user.EncryptedPassword).Scan(&user.ID); err != nil {
		return err
	}
	return nil
}

// FindByUsername ...
func (repository *UserRepository) FindByUsername(username string) (*model.User, error) {
	user := &model.User{}
	if err := repository.store.db.QueryRow(
		"SELECT id, username, encrypted_password FROM users WHERE username=$1",
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.EncryptedPassword,
	); err != nil {
		return nil, err
	}
	return user, nil
}

// FindByID ...
func (repository *UserRepository) FindByID(id int) (*model.User, error) {
	user := &model.User{ID: id}
	if err := repository.store.db.QueryRow(
		"SELECT username, encrypted_password FROM users WHERE id=$1",
		id,
	).Scan(
		&user.Username,
		&user.EncryptedPassword,
	); err != nil {
		return nil, err
	}
	return user, nil
}
