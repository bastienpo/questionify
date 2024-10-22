package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Conversation struct {
	ID        int64     `json:"conversation_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Title     string    `json:"title"`
	UserID    int64     `json:"user_id"`
	History   []string  `json:"history"`
	Version   int32     `json:"version"`
}

type ConversationModel struct {
	DB *sql.DB
}

func (m ConversationModel) Insert(conversation *Conversation) error {
	query := `
		INSERT INTO conversations (title, user_id, history)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{conversation.Title, conversation.UserID, conversation.History}

	return m.DB.QueryRowContext(ctx, query, args...).Scan(
		&conversation.ID,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
	)
}

func (m ConversationModel) Get(id int64) (*Conversation, error) {
	query := `
		SELECT id, created_at, updated_at, title, user_id, history, version
		FROM conversations
		WHERE id = $1
	`

	var conversation Conversation

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&conversation.ID,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
		&conversation.Title,
		&conversation.UserID,
		&conversation.History,
		&conversation.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &conversation, nil
}

func (m ConversationModel) Update(conversation *Conversation) error {
	query := `
		UPDATE conversations
		SET title = $1, user_id = $2, history = $3, version = version + 1
		WHERE id = $4 AND version = $5
		RETURNING version
	` // Avoid data race with version (optimistic locking)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		conversation.Title,
		conversation.UserID,
		conversation.History,
		conversation.ID,
		conversation.Version,
	}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&conversation.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m ConversationModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM conversations
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{id}

	result, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
