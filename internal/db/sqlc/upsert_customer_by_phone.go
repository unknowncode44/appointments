package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// UpsertCustomerByPhoneParams holds the arguments for UpsertCustomerByPhone.
type UpsertCustomerByPhoneParams struct {
	TenantID  uuid.UUID
	FirstName pgtype.Text
	LastName  pgtype.Text
	Phone     pgtype.Text
	Email     pgtype.Text
}

// UpsertCustomerByPhone atomically finds or creates a customer identified by
// (tenant_id, phone). It relies on the partial unique index on
// (tenant_id, phone) WHERE phone IS NOT NULL to stay race-safe under concurrent
// public bookings for the same number. On conflict the existing customer row
// (and its id) is preserved; names/email are filled in only when provided.
func (q *Queries) UpsertCustomerByPhone(ctx context.Context, arg UpsertCustomerByPhoneParams) (Customer, error) {
	const sql = `
INSERT INTO customers (tenant_id, first_name, last_name, phone, email)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (tenant_id, phone) WHERE phone IS NOT NULL
DO UPDATE SET
    first_name = COALESCE(EXCLUDED.first_name, customers.first_name),
    last_name  = COALESCE(EXCLUDED.last_name, customers.last_name),
    email      = COALESCE(EXCLUDED.email, customers.email),
    updated_at = now()
RETURNING ` + customerCols

	var i Customer
	err := q.db.QueryRow(ctx, sql, arg.TenantID, arg.FirstName, arg.LastName, arg.Phone, arg.Email).
		Scan(&i.ID, &i.TenantID, &i.FirstName, &i.LastName, &i.Notes, &i.CreatedAt, &i.UpdatedAt, &i.Phone, &i.Email)
	return i, err
}
