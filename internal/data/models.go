package data

import "database/sql"

type ModelStore struct {
	Users UserModel
}

func NewModelStore(db *sql.DB) *ModelStore {
	return &ModelStore{
		Users: UserModel{DB: db},
	}
}
