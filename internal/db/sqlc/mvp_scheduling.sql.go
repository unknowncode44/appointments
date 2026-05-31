package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateProviderAvailabilityParams struct {
	ProviderID uuid.UUID
	Weekday    int16
	StartTime  string
	EndTime    string
}

func (q *Queries) CreateProviderAvailability(ctx context.Context, arg CreateProviderAvailabilityParams) (ProviderAvailability, error) {
	var i ProviderAvailability
	err := q.db.QueryRow(ctx, "INSERT INTO provider_availability (provider_id, weekday, start_time, end_time) VALUES ($1, $2, $3, $4) RETURNING id, provider_id, weekday, start_time, end_time", arg.ProviderID, arg.Weekday, arg.StartTime, arg.EndTime).Scan(&i.ID, &i.ProviderID, &i.Weekday, &i.StartTime, &i.EndTime)
	return i, err
}

func (q *Queries) ListProviderAvailability(ctx context.Context, providerID uuid.UUID) ([]ProviderAvailability, error) {
	rows, err := q.db.Query(ctx, "SELECT id, provider_id, weekday, start_time, end_time FROM provider_availability WHERE provider_id = $1 ORDER BY weekday, start_time", providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ProviderAvailability{}
	for rows.Next() {
		var i ProviderAvailability
		if err := rows.Scan(&i.ID, &i.ProviderID, &i.Weekday, &i.StartTime, &i.EndTime); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

type CreateProviderExceptionParams struct {
	ProviderID uuid.UUID
	StartAt    time.Time
	EndAt      time.Time
	Reason     pgtype.Text
}

func (q *Queries) CreateProviderException(ctx context.Context, arg CreateProviderExceptionParams) (ProviderException, error) {
	var i ProviderException
	err := q.db.QueryRow(ctx, "INSERT INTO provider_exceptions (provider_id, start_at, end_at, reason) VALUES ($1, $2, $3, $4) RETURNING id, provider_id, start_at, end_at, reason, created_at", arg.ProviderID, arg.StartAt, arg.EndAt, arg.Reason).Scan(&i.ID, &i.ProviderID, &i.StartAt, &i.EndAt, &i.Reason, &i.CreatedAt)
	return i, err
}

func (q *Queries) ListProviderExceptions(ctx context.Context, providerID uuid.UUID) ([]ProviderException, error) {
	rows, err := q.db.Query(ctx, "SELECT id, provider_id, start_at, end_at, reason, created_at FROM provider_exceptions WHERE provider_id = $1 ORDER BY start_at DESC", providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProviderExceptions(rows)
}

type ListProviderExceptionsBetweenParams struct {
	ProviderID uuid.UUID
	StartAt    time.Time
	EndAt      time.Time
}

func (q *Queries) ListProviderExceptionsBetween(ctx context.Context, arg ListProviderExceptionsBetweenParams) ([]ProviderException, error) {
	rows, err := q.db.Query(ctx, "SELECT id, provider_id, start_at, end_at, reason, created_at FROM provider_exceptions WHERE provider_id = $1 AND start_at < $3 AND end_at > $2 ORDER BY start_at", arg.ProviderID, arg.StartAt, arg.EndAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProviderExceptions(rows)
}

type providerExceptionRows interface {
	Next() bool
	Scan(...interface{}) error
	Err() error
}

func scanProviderExceptions(rows providerExceptionRows) ([]ProviderException, error) {
	items := []ProviderException{}
	for rows.Next() {
		var i ProviderException
		if err := rows.Scan(&i.ID, &i.ProviderID, &i.StartAt, &i.EndAt, &i.Reason, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

type CreateAppointmentSlotParams struct {
	TenantID   uuid.UUID
	ProviderID uuid.UUID
	StartAt    time.Time
	EndAt      time.Time
}

func (q *Queries) CreateAppointmentSlot(ctx context.Context, arg CreateAppointmentSlotParams) (AppointmentSlot, error) {
	var i AppointmentSlot
	err := q.db.QueryRow(ctx, "INSERT INTO appointment_slots (tenant_id, provider_id, start_at, end_at, status) SELECT $1, $2, $3, $4, 'available' WHERE NOT EXISTS (SELECT 1 FROM appointment_slots WHERE provider_id = $2 AND start_at = $3 AND end_at = $4) RETURNING id, tenant_id, provider_id, start_at, end_at, status, appointment_id, created_at", arg.TenantID, arg.ProviderID, arg.StartAt, arg.EndAt).Scan(&i.ID, &i.TenantID, &i.ProviderID, &i.StartAt, &i.EndAt, &i.Status, &i.AppointmentID, &i.CreatedAt)
	return i, err
}

type ListAvailableSlotsParams struct {
	TenantID   uuid.UUID
	ProviderID *uuid.UUID
	StartAt    time.Time
	EndAt      time.Time
	ServiceID  *uuid.UUID
	Limit      int32
	Offset     int32
}

func (q *Queries) ListAvailableSlots(ctx context.Context, arg ListAvailableSlotsParams) ([]AppointmentSlot, error) {
	rows, err := q.db.Query(ctx, "SELECT s.id, s.tenant_id, s.provider_id, s.start_at, s.end_at, s.status, s.appointment_id, s.created_at FROM appointment_slots s WHERE s.tenant_id = $1 AND ($2::uuid IS NULL OR s.provider_id = $2) AND s.status = 'available' AND s.start_at >= $3 AND s.start_at < $4 AND ($5::uuid IS NULL OR EXISTS (SELECT 1 FROM provider_services ps WHERE ps.provider_id = s.provider_id AND ps.service_id = $5)) ORDER BY s.start_at LIMIT $6 OFFSET $7", arg.TenantID, arg.ProviderID, arg.StartAt, arg.EndAt, arg.ServiceID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []AppointmentSlot{}
	for rows.Next() {
		var i AppointmentSlot
		if err := rows.Scan(&i.ID, &i.TenantID, &i.ProviderID, &i.StartAt, &i.EndAt, &i.Status, &i.AppointmentID, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}
