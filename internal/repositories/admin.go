package repositories

import (
	"context"

	"github.com/google/uuid"
	db "github.com/mousav1/ticket/internal/db/sqlc"
)

type AdminRepository interface {
	ListTenants(context.Context, db.ListTenantsParams) ([]db.Tenant, int64, error)
	GetTenant(context.Context, uuid.UUID) (db.Tenant, error)
	CreateTenant(context.Context, db.CreateTenantParams) (db.Tenant, error)
	UpdateTenant(context.Context, db.UpdateTenantParams) (db.Tenant, error)
	DeactivateTenant(context.Context, uuid.UUID) (db.Tenant, error)
	ListProviders(context.Context, db.ListProvidersParams) ([]db.Provider, int64, error)
	GetProvider(context.Context, uuid.UUID) (db.Provider, error)
	CreateProvider(context.Context, db.CreateProviderParams) (db.Provider, error)
	UpdateProvider(context.Context, db.UpdateProviderParams) (db.Provider, error)
	DeactivateProvider(context.Context, uuid.UUID) (db.Provider, error)
	ListServices(context.Context, db.ListServicesParams) ([]db.Service, int64, error)
	GetService(context.Context, uuid.UUID) (db.Service, error)
	CreateService(context.Context, db.CreateServiceParams) (db.Service, error)
	UpdateService(context.Context, db.UpdateServiceParams) (db.Service, error)
	DeactivateService(context.Context, uuid.UUID) (db.Service, error)
	ListCustomers(context.Context, db.ListCustomersParams) ([]db.Customer, int64, error)
	GetCustomer(context.Context, uuid.UUID) (db.Customer, error)
	CreateCustomer(context.Context, db.CreateCustomerParams) (db.Customer, error)
	UpdateCustomer(context.Context, db.UpdateCustomerParams) (db.Customer, error)
}

type adminRepository struct{ store *db.Store }

func NewAdminRepository(store *db.Store) AdminRepository {
	return &adminRepository{store: store}
}

func (r *adminRepository) ListTenants(ctx context.Context, arg db.ListTenantsParams) ([]db.Tenant, int64, error) {
	items, err := r.store.ListTenants(ctx, arg)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.store.CountTenants(ctx, arg)
	return items, total, err
}

func (r *adminRepository) GetTenant(ctx context.Context, id uuid.UUID) (db.Tenant, error) {
	return r.store.GetTenant(ctx, id)
}
func (r *adminRepository) CreateTenant(ctx context.Context, arg db.CreateTenantParams) (db.Tenant, error) {
	return r.store.CreateTenant(ctx, arg)
}
func (r *adminRepository) UpdateTenant(ctx context.Context, arg db.UpdateTenantParams) (db.Tenant, error) {
	return r.store.UpdateTenant(ctx, arg)
}
func (r *adminRepository) DeactivateTenant(ctx context.Context, id uuid.UUID) (db.Tenant, error) {
	return r.store.DeactivateTenant(ctx, id)
}
func (r *adminRepository) ListProviders(ctx context.Context, arg db.ListProvidersParams) ([]db.Provider, int64, error) {
	items, err := r.store.ListProviders(ctx, arg)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.store.CountProviders(ctx, arg)
	return items, total, err
}
func (r *adminRepository) GetProvider(ctx context.Context, id uuid.UUID) (db.Provider, error) {
	return r.store.GetProvider(ctx, id)
}
func (r *adminRepository) CreateProvider(ctx context.Context, arg db.CreateProviderParams) (db.Provider, error) {
	return r.store.CreateProvider(ctx, arg)
}
func (r *adminRepository) UpdateProvider(ctx context.Context, arg db.UpdateProviderParams) (db.Provider, error) {
	return r.store.UpdateProvider(ctx, arg)
}
func (r *adminRepository) DeactivateProvider(ctx context.Context, id uuid.UUID) (db.Provider, error) {
	return r.store.DeactivateProvider(ctx, id)
}
func (r *adminRepository) ListServices(ctx context.Context, arg db.ListServicesParams) ([]db.Service, int64, error) {
	items, err := r.store.ListServices(ctx, arg)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.store.CountServices(ctx, arg)
	return items, total, err
}
func (r *adminRepository) GetService(ctx context.Context, id uuid.UUID) (db.Service, error) {
	return r.store.GetService(ctx, id)
}
func (r *adminRepository) CreateService(ctx context.Context, arg db.CreateServiceParams) (db.Service, error) {
	return r.store.CreateService(ctx, arg)
}
func (r *adminRepository) UpdateService(ctx context.Context, arg db.UpdateServiceParams) (db.Service, error) {
	return r.store.UpdateService(ctx, arg)
}
func (r *adminRepository) DeactivateService(ctx context.Context, id uuid.UUID) (db.Service, error) {
	return r.store.DeactivateService(ctx, id)
}
func (r *adminRepository) ListCustomers(ctx context.Context, arg db.ListCustomersParams) ([]db.Customer, int64, error) {
	items, err := r.store.ListCustomers(ctx, arg)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.store.CountCustomers(ctx, arg)
	return items, total, err
}
func (r *adminRepository) GetCustomer(ctx context.Context, id uuid.UUID) (db.Customer, error) {
	return r.store.GetCustomer(ctx, id)
}
func (r *adminRepository) CreateCustomer(ctx context.Context, arg db.CreateCustomerParams) (db.Customer, error) {
	return r.store.CreateCustomer(ctx, arg)
}
func (r *adminRepository) UpdateCustomer(ctx context.Context, arg db.UpdateCustomerParams) (db.Customer, error) {
	return r.store.UpdateCustomer(ctx, arg)
}
