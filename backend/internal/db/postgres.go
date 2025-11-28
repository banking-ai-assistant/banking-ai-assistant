package db

import (
	"database/sql"

	"banking-ai-assistant/internal/tickets"
)

type Repositories struct {
	Tickets tickets.Repository
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Tickets: tickets.NewPostgresRepo(db),
	}
}
