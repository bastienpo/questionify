package data

import "database/sql"

type ModelStore struct {
	Users         UserModel
	Conversations ConversationModel
}

func NewModelStore(db *sql.DB) *ModelStore {
	return &ModelStore{
		Users:         UserModel{DB: db},
		Conversations: ConversationModel{DB: db},
	}
}
