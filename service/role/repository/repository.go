package repository

import (
	"github.com/jmoiron/sqlx"
)

// Repositories is a struct containing role command and query repositories.
type Repositories interface {
	RoleQueryRepository
	RoleCommandRepository
}

type repositories struct {
	RoleQueryRepository
	RoleCommandRepository
}

// NewRepositories creates a new instance of Repositories with the provided database
// handle. This repository is responsible for executing command and query operations
// related to role records in the database.
//
// Parameters:
//   - db: A pointer to sqlx.DB containing the database handle.
//
// Returns:
//   - A pointer to the newly created Repositories instance.
func NewRepositories(db *sqlx.DB) Repositories {
	return &repositories{
		RoleQueryRepository:   NewRoleQueryRepository(db),
		RoleCommandRepository: NewRoleCommandRepository(db),
	}
}
