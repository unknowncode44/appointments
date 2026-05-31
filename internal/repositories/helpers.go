package repositories

import "github.com/jackc/pgx/v5/pgtype"

func Text(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func TextPtr(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}
