package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/mousav1/ticket/internal/api/dto"
	db "github.com/mousav1/ticket/internal/db/sqlc"
	"github.com/mousav1/ticket/internal/platform/pagination"
	"github.com/mousav1/ticket/internal/repositories"
)

type AdminService interface {
	ListTenants(context.Context, string, *bool, pagination.Page) (pagination.Response[dto.TenantResponse], error)
	GetTenant(context.Context, uuid.UUID) (dto.TenantResponse, error)
	CreateTenant(context.Context, dto.TenantRequest) (dto.TenantResponse, error)
	UpdateTenant(context.Context, uuid.UUID, dto.TenantRequest) (dto.TenantResponse, error)
	DeactivateTenant(context.Context, uuid.UUID) (dto.TenantResponse, error)
	ListProviders(context.Context, uuid.UUID, string, *bool, pagination.Page) (pagination.Response[dto.ProviderResponse], error)
	GetProvider(context.Context, uuid.UUID) (dto.ProviderResponse, error)
	CreateProvider(context.Context, dto.ProviderRequest) (dto.ProviderResponse, error)
	UpdateProvider(context.Context, uuid.UUID, dto.ProviderUpdateRequest) (dto.ProviderResponse, error)
	DeactivateProvider(context.Context, uuid.UUID) (dto.ProviderResponse, error)
	ListServices(context.Context, uuid.UUID, string, *bool, pagination.Page) (pagination.Response[dto.ServiceResponse], error)
	GetService(context.Context, uuid.UUID) (dto.ServiceResponse, error)
	CreateService(context.Context, dto.ServiceRequest) (dto.ServiceResponse, error)
	UpdateService(context.Context, uuid.UUID, dto.ServiceUpdateRequest) (dto.ServiceResponse, error)
	DeactivateService(context.Context, uuid.UUID) (dto.ServiceResponse, error)
	ListCustomers(context.Context, uuid.UUID, string, pagination.Page) (pagination.Response[dto.CustomerResponse], error)
	GetCustomer(context.Context, uuid.UUID) (dto.CustomerResponse, error)
	CreateCustomer(context.Context, dto.CustomerRequest) (dto.CustomerResponse, error)
	UpdateCustomer(context.Context, uuid.UUID, dto.CustomerUpdateRequest) (dto.CustomerResponse, error)
	ListTenantChannels(context.Context, uuid.UUID, string, *bool, pagination.Page) (pagination.Response[dto.TenantChannelResponse], error)
	GetTenantChannel(context.Context, uuid.UUID) (dto.TenantChannelResponse, error)
	CreateTenantChannel(context.Context, dto.TenantChannelRequest) (dto.TenantChannelResponse, error)
	UpdateTenantChannel(context.Context, uuid.UUID, dto.TenantChannelUpdateRequest) (dto.TenantChannelResponse, error)
	DeactivateTenantChannel(context.Context, uuid.UUID) (dto.TenantChannelResponse, error)
}

type adminService struct{ repo repositories.AdminRepository }

func NewAdminService(repo repositories.AdminRepository) AdminService {
	return &adminService{repo: repo}
}

func (s *adminService) ListTenants(ctx context.Context, search string, active *bool, p pagination.Page) (pagination.Response[dto.TenantResponse], error) {
	items, total, err := s.repo.ListTenants(ctx, db.ListTenantsParams{Search: search, Active: active, Limit: p.PageSize, Offset: p.Offset})
	return paged(items, total, p, mapTenant), err
}
func (s *adminService) GetTenant(ctx context.Context, id uuid.UUID) (dto.TenantResponse, error) {
	item, err := s.repo.GetTenant(ctx, id)
	return mapTenant(item), err
}
func (s *adminService) CreateTenant(ctx context.Context, req dto.TenantRequest) (dto.TenantResponse, error) {
	item, err := s.repo.CreateTenant(ctx, db.CreateTenantParams{Name: req.Name, Timezone: req.Timezone})
	return mapTenant(item), err
}
func (s *adminService) UpdateTenant(ctx context.Context, id uuid.UUID, req dto.TenantRequest) (dto.TenantResponse, error) {
	item, err := s.repo.UpdateTenant(ctx, db.UpdateTenantParams{ID: id, Name: req.Name, Timezone: req.Timezone})
	return mapTenant(item), err
}
func (s *adminService) DeactivateTenant(ctx context.Context, id uuid.UUID) (dto.TenantResponse, error) {
	item, err := s.repo.DeactivateTenant(ctx, id)
	return mapTenant(item), err
}
func (s *adminService) ListProviders(ctx context.Context, tenantID uuid.UUID, search string, active *bool, p pagination.Page) (pagination.Response[dto.ProviderResponse], error) {
	items, total, err := s.repo.ListProviders(ctx, db.ListProvidersParams{TenantID: tenantID, Search: search, Active: active, Limit: p.PageSize, Offset: p.Offset})
	return paged(items, total, p, mapProvider), err
}
func (s *adminService) GetProvider(ctx context.Context, id uuid.UUID) (dto.ProviderResponse, error) {
	item, err := s.repo.GetProvider(ctx, id)
	return mapProvider(item), err
}
func (s *adminService) CreateProvider(ctx context.Context, req dto.ProviderRequest) (dto.ProviderResponse, error) {
	item, err := s.repo.CreateProvider(ctx, db.CreateProviderParams{TenantID: req.TenantID, Name: req.Name})
	return mapProvider(item), err
}
func (s *adminService) UpdateProvider(ctx context.Context, id uuid.UUID, req dto.ProviderUpdateRequest) (dto.ProviderResponse, error) {
	item, err := s.repo.UpdateProvider(ctx, db.UpdateProviderParams{ID: id, Name: req.Name})
	return mapProvider(item), err
}
func (s *adminService) DeactivateProvider(ctx context.Context, id uuid.UUID) (dto.ProviderResponse, error) {
	item, err := s.repo.DeactivateProvider(ctx, id)
	return mapProvider(item), err
}
func (s *adminService) ListServices(ctx context.Context, tenantID uuid.UUID, search string, active *bool, p pagination.Page) (pagination.Response[dto.ServiceResponse], error) {
	items, total, err := s.repo.ListServices(ctx, db.ListServicesParams{TenantID: tenantID, Search: search, Active: active, Limit: p.PageSize, Offset: p.Offset})
	return paged(items, total, p, mapService), err
}
func (s *adminService) GetService(ctx context.Context, id uuid.UUID) (dto.ServiceResponse, error) {
	item, err := s.repo.GetService(ctx, id)
	return mapService(item), err
}
func (s *adminService) CreateService(ctx context.Context, req dto.ServiceRequest) (dto.ServiceResponse, error) {
	item, err := s.repo.CreateService(ctx, db.CreateServiceParams{TenantID: req.TenantID, Name: req.Name, DurationMinutes: req.DurationMinutes})
	return mapService(item), err
}
func (s *adminService) UpdateService(ctx context.Context, id uuid.UUID, req dto.ServiceUpdateRequest) (dto.ServiceResponse, error) {
	item, err := s.repo.UpdateService(ctx, db.UpdateServiceParams{ID: id, Name: req.Name, DurationMinutes: req.DurationMinutes})
	return mapService(item), err
}
func (s *adminService) DeactivateService(ctx context.Context, id uuid.UUID) (dto.ServiceResponse, error) {
	item, err := s.repo.DeactivateService(ctx, id)
	return mapService(item), err
}
func (s *adminService) ListCustomers(ctx context.Context, tenantID uuid.UUID, search string, p pagination.Page) (pagination.Response[dto.CustomerResponse], error) {
	items, total, err := s.repo.ListCustomers(ctx, db.ListCustomersParams{TenantID: tenantID, Search: search, Limit: p.PageSize, Offset: p.Offset})
	return paged(items, total, p, mapCustomer), err
}
func (s *adminService) GetCustomer(ctx context.Context, id uuid.UUID) (dto.CustomerResponse, error) {
	item, err := s.repo.GetCustomer(ctx, id)
	return mapCustomer(item), err
}
func (s *adminService) CreateCustomer(ctx context.Context, req dto.CustomerRequest) (dto.CustomerResponse, error) {
	item, err := s.repo.CreateCustomer(ctx, db.CreateCustomerParams{TenantID: req.TenantID, FirstName: repositories.Text(req.FirstName), LastName: repositories.Text(req.LastName), Notes: repositories.Text(req.Notes)})
	return mapCustomer(item), err
}
func (s *adminService) UpdateCustomer(ctx context.Context, id uuid.UUID, req dto.CustomerUpdateRequest) (dto.CustomerResponse, error) {
	item, err := s.repo.UpdateCustomer(ctx, db.UpdateCustomerParams{ID: id, FirstName: repositories.Text(req.FirstName), LastName: repositories.Text(req.LastName), Notes: repositories.Text(req.Notes)})
	return mapCustomer(item), err
}
func (s *adminService) ListTenantChannels(ctx context.Context, tenantID uuid.UUID, channelType string, active *bool, p pagination.Page) (pagination.Response[dto.TenantChannelResponse], error) {
	items, total, err := s.repo.ListTenantChannels(ctx, db.ListTenantChannelsParams{TenantID: tenantID, ChannelType: channelType, Active: active, Limit: p.PageSize, Offset: p.Offset})
	return paged(items, total, p, mapTenantChannel), err
}
func (s *adminService) GetTenantChannel(ctx context.Context, id uuid.UUID) (dto.TenantChannelResponse, error) {
	item, err := s.repo.GetTenantChannel(ctx, id)
	return mapTenantChannel(item), err
}
func (s *adminService) CreateTenantChannel(ctx context.Context, req dto.TenantChannelRequest) (dto.TenantChannelResponse, error) {
	item, err := s.repo.CreateTenantChannel(ctx, db.CreateTenantChannelParams{TenantID: req.TenantID, ChannelType: req.ChannelType, ExternalID: req.ExternalID, ExternalKey: repositories.Text(req.ExternalKey)})
	return mapTenantChannel(item), err
}
func (s *adminService) UpdateTenantChannel(ctx context.Context, id uuid.UUID, req dto.TenantChannelUpdateRequest) (dto.TenantChannelResponse, error) {
	item, err := s.repo.UpdateTenantChannel(ctx, db.UpdateTenantChannelParams{ID: id, ChannelType: req.ChannelType, ExternalID: req.ExternalID, ExternalKey: repositories.Text(req.ExternalKey), Active: req.Active})
	return mapTenantChannel(item), err
}
func (s *adminService) DeactivateTenantChannel(ctx context.Context, id uuid.UUID) (dto.TenantChannelResponse, error) {
	item, err := s.repo.DeactivateTenantChannel(ctx, id)
	return mapTenantChannel(item), err
}

func paged[I any, O any](items []I, total int64, p pagination.Page, mapper func(I) O) pagination.Response[O] {
	data := make([]O, 0, len(items))
	for _, item := range items {
		data = append(data, mapper(item))
	}
	return pagination.Response[O]{Data: data, Page: p.Page, PageSize: p.PageSize, Total: total}
}
