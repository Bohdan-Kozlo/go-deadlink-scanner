package scanner

import db "go-deadlink-scanner/internal/database/sqlc"

type Service struct {
	queries *db.Queries
}

func NewService(queries *db.Queries) *Service {
	return &Service{
		queries: queries,
	}
}
