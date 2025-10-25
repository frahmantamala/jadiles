package postgresql

import (
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewChildRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}
