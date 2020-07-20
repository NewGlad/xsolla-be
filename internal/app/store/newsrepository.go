package store

import (
	"fmt"

	"github.com/NewGlad/xsolla-be/internal/app/model"
)

// NewsRepository ...
type NewsRepository struct {
	store *Store
}

// Create ...
func (repository *NewsRepository) Create(news *model.News) error {
	if err := news.Validate(); err != nil {
		return err
	}
	if err := repository.store.db.QueryRow(
		"INSERT INTO news(content, author_id, likes) VALUES ($1, $2, 0) RETURNING id",
		news.Content, news.AuthorID).Scan(&news.ID); err != nil {
		return err
	}
	return nil
}

// GetTop ...
func (repository *NewsRepository) GetTop(top int) ([]*model.News, error) {
	rows, err := repository.store.db.Query(
		`SELECT id, content, author_id, likes
		FROM news
		ORDER BY likes DESC
		LIMIT $1`,
		top,
	)
	fmt.Printf("rows %v", rows)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	newsSlice := make([]*model.News, 0)
	for rows.Next() {
		news := &model.News{}
		if err := rows.Scan(&news.ID, &news.Content, &news.AuthorID, &news.Likes); err != nil {
			return nil, err
		}
		fmt.Printf("news %v", news)
		newsSlice = append(newsSlice, news)
		fmt.Printf("slice %v", newsSlice)
	}
	return newsSlice, nil
}

// FindByID ...
func (repository *NewsRepository) FindByID(ID int) (*model.News, error) {
	news := &model.News{ID: ID}
	if err := repository.store.db.QueryRow(
		"SELECT author_id, content, likes FROM news WHERE id=$1",
		news.ID,
	).Scan(
		&news.AuthorID,
		&news.Content,
		&news.Likes,
	); err != nil {
		return nil, err
	}
	return news, nil
}

// AddLike ...
func (repository *NewsRepository) AddLike(newsID int, userID int) error {
	tx, err := repository.store.db.Begin()
	defer tx.Rollback()
	if err != nil {
		return err
	}
	if _, err := tx.Exec("INSERT INTO likes(news_id, user_id) VALUES ($1, $2)", newsID, userID); err != nil {
		return fmt.Errorf("News with id '%d' does not exists or you can't like it twice", newsID)
	}
	if _, err := tx.Exec("UPDATE news SET likes=likes+1 WHERE id = $1", newsID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// RemoveLike ...
func (repository *NewsRepository) RemoveLike(newsID int, userID int) error {
	tx, err := repository.store.db.Begin()
	defer tx.Rollback()
	if err != nil {
		return err
	}
	result, err := tx.Exec("DELETE FROM likes WHERE user_id=$1 AND news_id=$2", userID, newsID)
	affectedRows, resultError := result.RowsAffected()
	if err != nil || affectedRows != 1 || resultError != nil {
		return fmt.Errorf("News with id '%d' not exists or you have no liked it", newsID)
	}
	if _, err := tx.Exec("UPDATE news SET likes=likes-1 WHERE id = $1;", newsID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
