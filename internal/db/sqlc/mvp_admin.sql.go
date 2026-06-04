package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const tenantCols = "id, name, timezone, active, created_at, updated_at"
const providerCols = "id, tenant_id, name, active, created_at, updated_at"
const serviceCols = "id, tenant_id, name, duration_minutes, active, created_at, updated_at"
const customerCols = "id, tenant_id, first_name, last_name, notes, created_at, updated_at"
const tenantChannelCols = "id, tenant_id, channel_type, external_id, external_key, active, created_at"

type ListTenantsParams struct {
	Search string
	Active *bool
	Limit  int32
	Offset int32
}

func (q *Queries) ListTenants(ctx context.Context, arg ListTenantsParams) ([]Tenant, error) {
	rows, err := q.db.Query(ctx, "SELECT "+tenantCols+" FROM tenants WHERE ($1::text = '' OR name ILIKE '%' || $1 || '%') AND ($2::boolean IS NULL OR active = $2) ORDER BY created_at DESC LIMIT $3 OFFSET $4", arg.Search, arg.Active, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Tenant{}
	for rows.Next() {
		var i Tenant
		if err := rows.Scan(&i.ID, &i.Name, &i.Timezone, &i.Active, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (q *Queries) CountTenants(ctx context.Context, arg ListTenantsParams) (int64, error) {
	var count int64
	err := q.db.QueryRow(ctx, "SELECT count(*) FROM tenants WHERE ($1::text = '' OR name ILIKE '%' || $1 || '%') AND ($2::boolean IS NULL OR active = $2)", arg.Search, arg.Active).Scan(&count)
	return count, err
}

func (q *Queries) GetTenant(ctx context.Context, id uuid.UUID) (Tenant, error) {
	var i Tenant
	err := q.db.QueryRow(ctx, "SELECT "+tenantCols+" FROM tenants WHERE id = $1", id).Scan(&i.ID, &i.Name, &i.Timezone, &i.Active, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

type CreateTenantParams struct {
	Name     string
	Timezone string
}

func (q *Queries) CreateTenant(ctx context.Context, arg CreateTenantParams) (Tenant, error) {
	var i Tenant
	err := q.db.QueryRow(ctx, "INSERT INTO tenants (name, timezone) VALUES ($1, $2) RETURNING "+tenantCols, arg.Name, arg.Timezone).Scan(&i.ID, &i.Name, &i.Timezone, &i.Active, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

type UpdateTenantParams struct {
	ID       uuid.UUID
	Name     string
	Timezone string
}

func (q *Queries) UpdateTenant(ctx context.Context, arg UpdateTenantParams) (Tenant, error) {
	var i Tenant
	err := q.db.QueryRow(ctx, "UPDATE tenants SET name = $2, timezone = $3, updated_at = now() WHERE id = $1 RETURNING "+tenantCols, arg.ID, arg.Name, arg.Timezone).Scan(&i.ID, &i.Name, &i.Timezone, &i.Active, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

func (q *Queries) DeactivateTenant(ctx context.Context, id uuid.UUID) (Tenant, error) {
	var i Tenant
	err := q.db.QueryRow(ctx, "UPDATE tenants SET active = false, updated_at = now() WHERE id = $1 RETURNING "+tenantCols, id).Scan(&i.ID, &i.Name, &i.Timezone, &i.Active, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

type ListProvidersParams struct {
	TenantID uuid.UUID
	Search   string
	Active   *bool
	Limit    int32
	Offset   int32
}

func (q *Queries) ListProviders(ctx context.Context, arg ListProvidersParams) ([]Provider, error) {
	rows, err := q.db.Query(ctx, "SELECT "+providerCols+" FROM providers WHERE tenant_id = $1 AND ($2::text = '' OR name ILIKE '%' || $2 || '%') AND ($3::boolean IS NULL OR active = $3) ORDER BY created_at DESC LIMIT $4 OFFSET $5", arg.TenantID, arg.Search, arg.Active, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Provider{}
	for rows.Next() {
		var i Provider
		if err := rows.Scan(&i.ID, &i.TenantID, &i.Name, &i.Active, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (q *Queries) CountProviders(ctx context.Context, arg ListProvidersParams) (int64, error) {
	var count int64
	err := q.db.QueryRow(ctx, "SELECT count(*) FROM providers WHERE tenant_id = $1 AND ($2::text = '' OR name ILIKE '%' || $2 || '%') AND ($3::boolean IS NULL OR active = $3)", arg.TenantID, arg.Search, arg.Active).Scan(&count)
	return count, err
}

func (q *Queries) GetProvider(ctx context.Context, id uuid.UUID) (Provider, error) {
	var i Provider
	err := q.db.QueryRow(ctx, "SELECT "+providerCols+" FROM providers WHERE id = $1", id).Scan(&i.ID, &i.TenantID, &i.Name, &i.Active, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

type CreateProviderParams struct {
	TenantID uuid.UUID
	Name     string
}

func (q *Queries) CreateProvider(ctx context.Context, arg CreateProviderParams) (Provider, error) {
	var i Provider
	err := q.db.QueryRow(ctx, "INSERT INTO providers (tenant_id, name) VALUES ($1, $2) RETURNING "+providerCols, arg.TenantID, arg.Name).Scan(&i.ID, &i.TenantID, &i.Name, &i.Active, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

type UpdateProviderParams struct {
	ID   uuid.UUID
	Name string
}

func (q *Queries) UpdateProvider(ctx context.Context, arg UpdateProviderParams) (Provider, error) {
	var i Provider
	err := q.db.QueryRow(ctx, "UPDATE providers SET name = $2, updated_at = now() WHERE id = $1 RETURNING "+providerCols, arg.ID, arg.Name).Scan(&i.ID, &i.TenantID, &i.Name, &i.Active, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

func (q *Queries) DeactivateProvider(ctx context.Context, id uuid.UUID) (Provider, error) {
	var i Provider
	err := q.db.QueryRow(ctx, "UPDATE providers SET active = false, updated_at = now() WHERE id = $1 RETURNING "+providerCols, id).Scan(&i.ID, &i.TenantID, &i.Name, &i.Active, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

type ListServicesParams = ListProvidersParams

func (q *Queries) ListServices(ctx context.Context, arg ListServicesParams) ([]Service, error) {
	rows, err := q.db.Query(ctx, "SELECT "+serviceCols+" FROM services WHERE tenant_id = $1 AND ($2::text = '' OR name ILIKE '%' || $2 || '%') AND ($3::boolean IS NULL OR active = $3) ORDER BY created_at DESC LIMIT $4 OFFSET $5", arg.TenantID, arg.Search, arg.Active, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Service{}
	for rows.Next() {
		var i Service
		if err := rows.Scan(&i.ID, &i.TenantID, &i.Name, &i.DurationMinutes, &i.Active, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (q *Queries) CountServices(ctx context.Context, arg ListServicesParams) (int64, error) {
	var count int64
	err := q.db.QueryRow(ctx, "SELECT count(*) FROM services WHERE tenant_id = $1 AND ($2::text = '' OR name ILIKE '%' || $2 || '%') AND ($3::boolean IS NULL OR active = $3)", arg.TenantID, arg.Search, arg.Active).Scan(&count)
	return count, err
}

func (q *Queries) GetService(ctx context.Context, id uuid.UUID) (Service, error) {
	var i Service
	err := q.db.QueryRow(ctx, "SELECT "+serviceCols+" FROM services WHERE id = $1", id).Scan(&i.ID, &i.TenantID, &i.Name, &i.DurationMinutes, &i.Active, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

type CreateServiceParams struct {
	TenantID        uuid.UUID
	Name            string
	DurationMinutes int32
}

func (q *Queries) CreateService(ctx context.Context, arg CreateServiceParams) (Service, error) {
	var i Service
	err := q.db.QueryRow(ctx, "INSERT INTO services (tenant_id, name, duration_minutes) VALUES ($1, $2, $3) RETURNING "+serviceCols, arg.TenantID, arg.Name, arg.DurationMinutes).Scan(&i.ID, &i.TenantID, &i.Name, &i.DurationMinutes, &i.Active, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

type UpdateServiceParams struct {
	ID              uuid.UUID
	Name            string
	DurationMinutes int32
}

func (q *Queries) UpdateService(ctx context.Context, arg UpdateServiceParams) (Service, error) {
	var i Service
	err := q.db.QueryRow(ctx, "UPDATE services SET name = $2, duration_minutes = $3, updated_at = now() WHERE id = $1 RETURNING "+serviceCols, arg.ID, arg.Name, arg.DurationMinutes).Scan(&i.ID, &i.TenantID, &i.Name, &i.DurationMinutes, &i.Active, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

func (q *Queries) DeactivateService(ctx context.Context, id uuid.UUID) (Service, error) {
	var i Service
	err := q.db.QueryRow(ctx, "UPDATE services SET active = false, updated_at = now() WHERE id = $1 RETURNING "+serviceCols, id).Scan(&i.ID, &i.TenantID, &i.Name, &i.DurationMinutes, &i.Active, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

type ListCustomersParams struct {
	TenantID uuid.UUID
	Search   string
	Limit    int32
	Offset   int32
}

func (q *Queries) ListCustomers(ctx context.Context, arg ListCustomersParams) ([]Customer, error) {
	rows, err := q.db.Query(ctx, "SELECT "+customerCols+" FROM customers WHERE tenant_id = $1 AND ($2::text = '' OR coalesce(first_name, '') ILIKE '%' || $2 || '%' OR coalesce(last_name, '') ILIKE '%' || $2 || '%') ORDER BY created_at DESC LIMIT $3 OFFSET $4", arg.TenantID, arg.Search, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Customer{}
	for rows.Next() {
		var i Customer
		if err := rows.Scan(&i.ID, &i.TenantID, &i.FirstName, &i.LastName, &i.Notes, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (q *Queries) CountCustomers(ctx context.Context, arg ListCustomersParams) (int64, error) {
	var count int64
	err := q.db.QueryRow(ctx, "SELECT count(*) FROM customers WHERE tenant_id = $1 AND ($2::text = '' OR coalesce(first_name, '') ILIKE '%' || $2 || '%' OR coalesce(last_name, '') ILIKE '%' || $2 || '%')", arg.TenantID, arg.Search).Scan(&count)
	return count, err
}

func (q *Queries) GetCustomer(ctx context.Context, id uuid.UUID) (Customer, error) {
	var i Customer
	err := q.db.QueryRow(ctx, "SELECT "+customerCols+" FROM customers WHERE id = $1", id).Scan(&i.ID, &i.TenantID, &i.FirstName, &i.LastName, &i.Notes, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

type CreateCustomerParams struct {
	TenantID  uuid.UUID
	FirstName pgtype.Text
	LastName  pgtype.Text
	Notes     pgtype.Text
}

func (q *Queries) CreateCustomer(ctx context.Context, arg CreateCustomerParams) (Customer, error) {
	var i Customer
	err := q.db.QueryRow(ctx, "INSERT INTO customers (tenant_id, first_name, last_name, notes) VALUES ($1, $2, $3, $4) RETURNING "+customerCols, arg.TenantID, arg.FirstName, arg.LastName, arg.Notes).Scan(&i.ID, &i.TenantID, &i.FirstName, &i.LastName, &i.Notes, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

type UpdateCustomerParams struct {
	ID        uuid.UUID
	FirstName pgtype.Text
	LastName  pgtype.Text
	Notes     pgtype.Text
}

func (q *Queries) UpdateCustomer(ctx context.Context, arg UpdateCustomerParams) (Customer, error) {
	var i Customer
	err := q.db.QueryRow(ctx, "UPDATE customers SET first_name = $2, last_name = $3, notes = $4, updated_at = now() WHERE id = $1 RETURNING "+customerCols, arg.ID, arg.FirstName, arg.LastName, arg.Notes).Scan(&i.ID, &i.TenantID, &i.FirstName, &i.LastName, &i.Notes, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

type ListTenantChannelsParams struct {
	TenantID    uuid.UUID
	ChannelType string
	Active      *bool
	Limit       int32
	Offset      int32
}

func (q *Queries) ListTenantChannels(ctx context.Context, arg ListTenantChannelsParams) ([]TenantChannel, error) {
	rows, err := q.db.Query(ctx, "SELECT "+tenantChannelCols+" FROM tenant_channels WHERE tenant_id = $1 AND ($2::text = '' OR channel_type = $2) AND ($3::boolean IS NULL OR active = $3) ORDER BY created_at DESC LIMIT $4 OFFSET $5", arg.TenantID, arg.ChannelType, arg.Active, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TenantChannel{}
	for rows.Next() {
		var i TenantChannel
		if err := rows.Scan(&i.ID, &i.TenantID, &i.ChannelType, &i.ExternalID, &i.ExternalKey, &i.Active, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (q *Queries) CountTenantChannels(ctx context.Context, arg ListTenantChannelsParams) (int64, error) {
	var count int64
	err := q.db.QueryRow(ctx, "SELECT count(*) FROM tenant_channels WHERE tenant_id = $1 AND ($2::text = '' OR channel_type = $2) AND ($3::boolean IS NULL OR active = $3)", arg.TenantID, arg.ChannelType, arg.Active).Scan(&count)
	return count, err
}

func (q *Queries) GetTenantChannel(ctx context.Context, id uuid.UUID) (TenantChannel, error) {
	var i TenantChannel
	err := q.db.QueryRow(ctx, "SELECT "+tenantChannelCols+" FROM tenant_channels WHERE id = $1", id).Scan(&i.ID, &i.TenantID, &i.ChannelType, &i.ExternalID, &i.ExternalKey, &i.Active, &i.CreatedAt)
	return i, err
}

type CreateTenantChannelParams struct {
	TenantID    uuid.UUID
	ChannelType string
	ExternalID  string
	ExternalKey pgtype.Text
}

func (q *Queries) CreateTenantChannel(ctx context.Context, arg CreateTenantChannelParams) (TenantChannel, error) {
	var i TenantChannel
	err := q.db.QueryRow(ctx, "INSERT INTO tenant_channels (tenant_id, channel_type, external_id, external_key) VALUES ($1, $2, $3, $4) RETURNING "+tenantChannelCols, arg.TenantID, arg.ChannelType, arg.ExternalID, arg.ExternalKey).Scan(&i.ID, &i.TenantID, &i.ChannelType, &i.ExternalID, &i.ExternalKey, &i.Active, &i.CreatedAt)
	return i, err
}

type UpdateTenantChannelParams struct {
	ID          uuid.UUID
	ChannelType string
	ExternalID  string
	ExternalKey pgtype.Text
	Active      bool
}

func (q *Queries) UpdateTenantChannel(ctx context.Context, arg UpdateTenantChannelParams) (TenantChannel, error) {
	var i TenantChannel
	err := q.db.QueryRow(ctx, "UPDATE tenant_channels SET channel_type = $2, external_id = $3, external_key = $4, active = $5 WHERE id = $1 RETURNING "+tenantChannelCols, arg.ID, arg.ChannelType, arg.ExternalID, arg.ExternalKey, arg.Active).Scan(&i.ID, &i.TenantID, &i.ChannelType, &i.ExternalID, &i.ExternalKey, &i.Active, &i.CreatedAt)
	return i, err
}

func (q *Queries) DeactivateTenantChannel(ctx context.Context, id uuid.UUID) (TenantChannel, error) {
	var i TenantChannel
	err := q.db.QueryRow(ctx, "UPDATE tenant_channels SET active = false WHERE id = $1 RETURNING "+tenantChannelCols, id).Scan(&i.ID, &i.TenantID, &i.ChannelType, &i.ExternalID, &i.ExternalKey, &i.Active, &i.CreatedAt)
	return i, err
}
