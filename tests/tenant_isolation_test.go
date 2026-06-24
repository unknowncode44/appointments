package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/unknowncode44/appointments/internal/api/dto"
	"github.com/unknowncode44/appointments/internal/api/response"
	db "github.com/unknowncode44/appointments/internal/db/sqlc"
	"github.com/unknowncode44/appointments/internal/services"
)

// errStubNotImpl is returned by stub methods that are not exercised by the tests.
var errStubNotImpl = errors.New("stub: not implemented")

// ─── Admin repository stub ────────────────────────────────────────────────────

// stubAdminRepo satisfies repositories.AdminRepository.
// Only the Get*/Update*/Deactivate* methods that are exercised by tests return
// real data; the rest return errStubNotImpl so we can detect accidental calls.
type stubAdminRepo struct {
	provider db.Provider
	service  db.Service
	customer db.Customer
	channel  db.TenantChannel
}

func (r *stubAdminRepo) ListTenants(_ context.Context, _ db.ListTenantsParams) ([]db.Tenant, int64, error) {
	return nil, 0, errStubNotImpl
}
func (r *stubAdminRepo) GetTenant(_ context.Context, _ uuid.UUID) (db.Tenant, error) {
	return db.Tenant{}, errStubNotImpl
}
func (r *stubAdminRepo) CreateTenant(_ context.Context, _ db.CreateTenantParams) (db.Tenant, error) {
	return db.Tenant{}, errStubNotImpl
}
func (r *stubAdminRepo) UpdateTenant(_ context.Context, _ db.UpdateTenantParams) (db.Tenant, error) {
	return db.Tenant{}, errStubNotImpl
}
func (r *stubAdminRepo) DeactivateTenant(_ context.Context, _ uuid.UUID) (db.Tenant, error) {
	return db.Tenant{}, errStubNotImpl
}
func (r *stubAdminRepo) ListProviders(_ context.Context, _ db.ListProvidersParams) ([]db.Provider, int64, error) {
	return nil, 0, errStubNotImpl
}
func (r *stubAdminRepo) GetProvider(_ context.Context, _ uuid.UUID) (db.Provider, error) {
	return r.provider, nil
}
func (r *stubAdminRepo) CreateProvider(_ context.Context, _ db.CreateProviderParams) (db.Provider, error) {
	return db.Provider{}, errStubNotImpl
}
func (r *stubAdminRepo) UpdateProvider(_ context.Context, _ db.UpdateProviderParams) (db.Provider, error) {
	return r.provider, nil
}
func (r *stubAdminRepo) DeactivateProvider(_ context.Context, _ uuid.UUID) (db.Provider, error) {
	return r.provider, nil
}
func (r *stubAdminRepo) ListServices(_ context.Context, _ db.ListServicesParams) ([]db.Service, int64, error) {
	return nil, 0, errStubNotImpl
}
func (r *stubAdminRepo) GetService(_ context.Context, _ uuid.UUID) (db.Service, error) {
	return r.service, nil
}
func (r *stubAdminRepo) CreateService(_ context.Context, _ db.CreateServiceParams) (db.Service, error) {
	return db.Service{}, errStubNotImpl
}
func (r *stubAdminRepo) UpdateService(_ context.Context, _ db.UpdateServiceParams) (db.Service, error) {
	return r.service, nil
}
func (r *stubAdminRepo) DeactivateService(_ context.Context, _ uuid.UUID) (db.Service, error) {
	return r.service, nil
}
func (r *stubAdminRepo) ListCustomers(_ context.Context, _ db.ListCustomersParams) ([]db.Customer, int64, error) {
	return nil, 0, errStubNotImpl
}
func (r *stubAdminRepo) GetCustomer(_ context.Context, _ uuid.UUID) (db.Customer, error) {
	return r.customer, nil
}
func (r *stubAdminRepo) CreateCustomer(_ context.Context, _ db.CreateCustomerParams) (db.Customer, error) {
	return db.Customer{}, errStubNotImpl
}
func (r *stubAdminRepo) UpdateCustomer(_ context.Context, _ db.UpdateCustomerParams) (db.Customer, error) {
	return r.customer, nil
}
func (r *stubAdminRepo) ListTenantChannels(_ context.Context, _ db.ListTenantChannelsParams) ([]db.TenantChannel, int64, error) {
	return nil, 0, errStubNotImpl
}
func (r *stubAdminRepo) GetTenantChannel(_ context.Context, _ uuid.UUID) (db.TenantChannel, error) {
	return r.channel, nil
}
func (r *stubAdminRepo) CreateTenantChannel(_ context.Context, _ db.CreateTenantChannelParams) (db.TenantChannel, error) {
	return db.TenantChannel{}, errStubNotImpl
}
func (r *stubAdminRepo) UpdateTenantChannel(_ context.Context, _ db.UpdateTenantChannelParams) (db.TenantChannel, error) {
	return r.channel, nil
}
func (r *stubAdminRepo) DeactivateTenantChannel(_ context.Context, _ uuid.UUID) (db.TenantChannel, error) {
	return r.channel, nil
}

// ─── Workflow repository stub ─────────────────────────────────────────────────

// stubWorkflowRepo satisfies repositories.WorkflowRepository.
// Store() returns nil because Get/GetThread do not invoke ExecTx.
type stubWorkflowRepo struct {
	appointment db.Appointment
	thread      db.ConversationThread
	customer    db.Customer
}

func (r *stubWorkflowRepo) Store() *db.Store { return nil }
func (r *stubWorkflowRepo) GetCustomer(_ context.Context, _ uuid.UUID) (db.Customer, error) {
	return r.customer, nil
}
func (r *stubWorkflowRepo) ListAppointments(_ context.Context, _ db.ListAppointmentsParams) ([]db.Appointment, int64, error) {
	return nil, 0, errStubNotImpl
}
func (r *stubWorkflowRepo) GetAppointment(_ context.Context, _ uuid.UUID) (db.Appointment, error) {
	return r.appointment, nil
}
func (r *stubWorkflowRepo) ListThreads(_ context.Context, _ db.ListConversationThreadsParams) ([]db.ConversationThread, error) {
	return nil, errStubNotImpl
}
func (r *stubWorkflowRepo) GetThread(_ context.Context, _ uuid.UUID) (db.ConversationThread, error) {
	return r.thread, nil
}
func (r *stubWorkflowRepo) ListMessages(_ context.Context, _ uuid.UUID) ([]db.ConversationMessage, error) {
	return nil, nil
}
func (r *stubWorkflowRepo) GetTenantChannelByExternalID(_ context.Context, _ string) (db.TenantChannel, error) {
	return db.TenantChannel{}, errStubNotImpl
}

// ─── Scheduling repository stub ───────────────────────────────────────────────

// stubSchedulingRepo satisfies repositories.SchedulingRepository.
// GetProvider returns the configured provider; the list/create methods return
// empty data so the scope checks can be exercised without a DB.
type stubSchedulingRepo struct {
	provider db.Provider
}

func (r *stubSchedulingRepo) GetProvider(_ context.Context, _ uuid.UUID) (db.Provider, error) {
	return r.provider, nil
}
func (r *stubSchedulingRepo) CreateAvailability(_ context.Context, _ db.CreateProviderAvailabilityParams) (db.ProviderAvailability, error) {
	return db.ProviderAvailability{}, nil
}
func (r *stubSchedulingRepo) ListAvailability(_ context.Context, _ uuid.UUID) ([]db.ProviderAvailability, error) {
	return nil, nil
}
func (r *stubSchedulingRepo) CreateException(_ context.Context, _ db.CreateProviderExceptionParams) (db.ProviderException, error) {
	return db.ProviderException{}, nil
}
func (r *stubSchedulingRepo) ListExceptions(_ context.Context, _ uuid.UUID) ([]db.ProviderException, error) {
	return nil, nil
}
func (r *stubSchedulingRepo) ListExceptionsBetween(_ context.Context, _ db.ListProviderExceptionsBetweenParams) ([]db.ProviderException, error) {
	return nil, nil
}
func (r *stubSchedulingRepo) CreateSlot(_ context.Context, _ db.CreateAppointmentSlotParams) (db.AppointmentSlot, error) {
	return db.AppointmentSlot{}, nil
}
func (r *stubSchedulingRepo) ListAvailableSlots(_ context.Context, _ db.ListAvailableSlotsParams) ([]db.AppointmentSlot, error) {
	return nil, nil
}

// ─── Provider isolation ───────────────────────────────────────────────────────

func TestTenantIsolation_GetProvider(t *testing.T) {
	tenantA, tenantB, providerID := uuid.New(), uuid.New(), uuid.New()

	svc := services.NewAdminService(&stubAdminRepo{
		provider: db.Provider{ID: providerID, TenantID: tenantA},
	})

	t.Run("cross-tenant returns ErrNotFound", func(t *testing.T) {
		_, err := svc.GetProvider(context.Background(), providerID, &tenantB)
		require.ErrorIs(t, err, response.ErrNotFound)
	})

	t.Run("own tenant succeeds", func(t *testing.T) {
		rsp, err := svc.GetProvider(context.Background(), providerID, &tenantA)
		require.NoError(t, err)
		require.Equal(t, providerID, rsp.ID)
	})

	t.Run("nil scope (admin) bypasses check", func(t *testing.T) {
		rsp, err := svc.GetProvider(context.Background(), providerID, nil)
		require.NoError(t, err)
		require.Equal(t, providerID, rsp.ID)
	})
}

func TestTenantIsolation_UpdateProvider(t *testing.T) {
	tenantA, tenantB, providerID := uuid.New(), uuid.New(), uuid.New()

	svc := services.NewAdminService(&stubAdminRepo{
		provider: db.Provider{ID: providerID, TenantID: tenantA},
	})
	req := dto.ProviderUpdateRequest{Name: "renamed"}

	t.Run("cross-tenant update returns ErrNotFound", func(t *testing.T) {
		_, err := svc.UpdateProvider(context.Background(), providerID, req, &tenantB)
		require.ErrorIs(t, err, response.ErrNotFound)
	})

	t.Run("own tenant update succeeds", func(t *testing.T) {
		_, err := svc.UpdateProvider(context.Background(), providerID, req, &tenantA)
		require.NoError(t, err)
	})

	t.Run("nil scope (admin) succeeds", func(t *testing.T) {
		_, err := svc.UpdateProvider(context.Background(), providerID, req, nil)
		require.NoError(t, err)
	})
}

func TestTenantIsolation_DeactivateProvider(t *testing.T) {
	tenantA, tenantB, providerID := uuid.New(), uuid.New(), uuid.New()

	svc := services.NewAdminService(&stubAdminRepo{
		provider: db.Provider{ID: providerID, TenantID: tenantA},
	})

	t.Run("cross-tenant deactivate returns ErrNotFound", func(t *testing.T) {
		_, err := svc.DeactivateProvider(context.Background(), providerID, &tenantB)
		require.ErrorIs(t, err, response.ErrNotFound)
	})

	t.Run("own tenant deactivate succeeds", func(t *testing.T) {
		_, err := svc.DeactivateProvider(context.Background(), providerID, &tenantA)
		require.NoError(t, err)
	})
}

// ─── Service isolation ────────────────────────────────────────────────────────

func TestTenantIsolation_GetService(t *testing.T) {
	tenantA, tenantB, serviceID := uuid.New(), uuid.New(), uuid.New()

	svc := services.NewAdminService(&stubAdminRepo{
		service: db.Service{ID: serviceID, TenantID: tenantA},
	})

	t.Run("cross-tenant returns ErrNotFound", func(t *testing.T) {
		_, err := svc.GetService(context.Background(), serviceID, &tenantB)
		require.ErrorIs(t, err, response.ErrNotFound)
	})

	t.Run("own tenant succeeds", func(t *testing.T) {
		rsp, err := svc.GetService(context.Background(), serviceID, &tenantA)
		require.NoError(t, err)
		require.Equal(t, serviceID, rsp.ID)
	})
}

func TestTenantIsolation_DeactivateService(t *testing.T) {
	tenantA, tenantB, serviceID := uuid.New(), uuid.New(), uuid.New()

	svc := services.NewAdminService(&stubAdminRepo{
		service: db.Service{ID: serviceID, TenantID: tenantA},
	})

	t.Run("cross-tenant deactivate returns ErrNotFound", func(t *testing.T) {
		_, err := svc.DeactivateService(context.Background(), serviceID, &tenantB)
		require.ErrorIs(t, err, response.ErrNotFound)
	})

	t.Run("own tenant deactivate succeeds", func(t *testing.T) {
		_, err := svc.DeactivateService(context.Background(), serviceID, &tenantA)
		require.NoError(t, err)
	})
}

// ─── Customer isolation ───────────────────────────────────────────────────────

func TestTenantIsolation_GetCustomer(t *testing.T) {
	tenantA, tenantB, customerID := uuid.New(), uuid.New(), uuid.New()

	svc := services.NewAdminService(&stubAdminRepo{
		customer: db.Customer{ID: customerID, TenantID: tenantA},
	})

	t.Run("cross-tenant returns ErrNotFound", func(t *testing.T) {
		_, err := svc.GetCustomer(context.Background(), customerID, &tenantB)
		require.ErrorIs(t, err, response.ErrNotFound)
	})

	t.Run("nil scope (admin) succeeds", func(t *testing.T) {
		rsp, err := svc.GetCustomer(context.Background(), customerID, nil)
		require.NoError(t, err)
		require.Equal(t, customerID, rsp.ID)
	})
}

// ─── TenantChannel isolation ──────────────────────────────────────────────────

func TestTenantIsolation_GetTenantChannel(t *testing.T) {
	tenantA, tenantB, channelID := uuid.New(), uuid.New(), uuid.New()

	svc := services.NewAdminService(&stubAdminRepo{
		channel: db.TenantChannel{ID: channelID, TenantID: tenantA},
	})

	t.Run("cross-tenant returns ErrNotFound", func(t *testing.T) {
		_, err := svc.GetTenantChannel(context.Background(), channelID, &tenantB)
		require.ErrorIs(t, err, response.ErrNotFound)
	})

	t.Run("own tenant succeeds", func(t *testing.T) {
		rsp, err := svc.GetTenantChannel(context.Background(), channelID, &tenantA)
		require.NoError(t, err)
		require.Equal(t, channelID, rsp.ID)
	})
}

// ─── Appointment isolation ────────────────────────────────────────────────────

func TestTenantIsolation_GetAppointment(t *testing.T) {
	tenantA, tenantB, apptID := uuid.New(), uuid.New(), uuid.New()

	svc := services.NewAppointmentService(&stubWorkflowRepo{
		appointment: db.Appointment{ID: apptID, TenantID: tenantA, Status: "confirmed"},
	})

	t.Run("cross-tenant returns ErrNotFound", func(t *testing.T) {
		_, err := svc.Get(context.Background(), apptID, &tenantB)
		require.ErrorIs(t, err, response.ErrNotFound)
	})

	t.Run("own tenant succeeds", func(t *testing.T) {
		rsp, err := svc.Get(context.Background(), apptID, &tenantA)
		require.NoError(t, err)
		require.Equal(t, apptID, rsp.ID)
	})

	t.Run("nil scope (admin) bypasses check", func(t *testing.T) {
		rsp, err := svc.Get(context.Background(), apptID, nil)
		require.NoError(t, err)
		require.Equal(t, apptID, rsp.ID)
	})
}

// ─── ConversationThread isolation ─────────────────────────────────────────────

func TestTenantIsolation_GetConversationThread(t *testing.T) {
	tenantA, tenantB, threadID := uuid.New(), uuid.New(), uuid.New()

	svc := services.NewConversationService(&stubWorkflowRepo{
		thread: db.ConversationThread{ID: threadID, TenantID: tenantA},
	})

	t.Run("cross-tenant returns ErrNotFound", func(t *testing.T) {
		_, err := svc.GetThread(context.Background(), threadID, &tenantB)
		require.ErrorIs(t, err, response.ErrNotFound)
	})

	t.Run("own tenant succeeds", func(t *testing.T) {
		rsp, err := svc.GetThread(context.Background(), threadID, &tenantA)
		require.NoError(t, err)
		require.Equal(t, threadID, rsp.ID)
	})
}

// ─── Scheduling isolation ─────────────────────────────────────────────────────

func TestTenantIsolation_CreateAvailability(t *testing.T) {
	tenantA, tenantB, providerID := uuid.New(), uuid.New(), uuid.New()

	svc := services.NewSchedulingService(&stubSchedulingRepo{
		provider: db.Provider{ID: providerID, TenantID: tenantA},
	})
	req := dto.AvailabilityRequest{Weekday: 1, StartTime: "09:00", EndTime: "17:00"}

	t.Run("cross-tenant provider returns ErrNotFound", func(t *testing.T) {
		_, err := svc.CreateAvailability(context.Background(), providerID, req, &tenantB)
		require.ErrorIs(t, err, response.ErrNotFound)
	})

	t.Run("own tenant succeeds", func(t *testing.T) {
		_, err := svc.CreateAvailability(context.Background(), providerID, req, &tenantA)
		require.NoError(t, err)
	})

	t.Run("nil scope (admin) bypasses check", func(t *testing.T) {
		_, err := svc.CreateAvailability(context.Background(), providerID, req, nil)
		require.NoError(t, err)
	})
}

func TestTenantIsolation_CreateException(t *testing.T) {
	tenantA, tenantB, providerID := uuid.New(), uuid.New(), uuid.New()

	svc := services.NewSchedulingService(&stubSchedulingRepo{
		provider: db.Provider{ID: providerID, TenantID: tenantA},
	})
	req := dto.ExceptionRequest{StartAt: time.Now(), EndAt: time.Now().Add(time.Hour)}

	t.Run("cross-tenant provider returns ErrNotFound", func(t *testing.T) {
		_, err := svc.CreateException(context.Background(), providerID, req, &tenantB)
		require.ErrorIs(t, err, response.ErrNotFound)
	})

	t.Run("own tenant succeeds", func(t *testing.T) {
		_, err := svc.CreateException(context.Background(), providerID, req, &tenantA)
		require.NoError(t, err)
	})

	t.Run("nil scope (admin) bypasses check", func(t *testing.T) {
		_, err := svc.CreateException(context.Background(), providerID, req, nil)
		require.NoError(t, err)
	})
}

func TestTenantIsolation_GenerateSlots(t *testing.T) {
	tenantA, tenantB, providerID := uuid.New(), uuid.New(), uuid.New()

	svc := services.NewSchedulingService(&stubSchedulingRepo{
		provider: db.Provider{ID: providerID, TenantID: tenantA},
	})
	req := dto.SlotGeneratorRequest{TenantID: tenantA, ProviderID: providerID, NumberOfWeeks: 1}

	t.Run("cross-tenant provider returns ErrNotFound", func(t *testing.T) {
		_, err := svc.GenerateSlots(context.Background(), req, &tenantB)
		require.ErrorIs(t, err, response.ErrNotFound)
	})

	t.Run("own tenant succeeds", func(t *testing.T) {
		_, err := svc.GenerateSlots(context.Background(), req, &tenantA)
		require.NoError(t, err)
	})

	t.Run("nil scope (admin) bypasses check", func(t *testing.T) {
		_, err := svc.GenerateSlots(context.Background(), req, nil)
		require.NoError(t, err)
	})
}

// ─── Conversation message isolation ───────────────────────────────────────────

func TestTenantIsolation_StoreMessage(t *testing.T) {
	tenantA, tenantB, customerID := uuid.New(), uuid.New(), uuid.New()

	svc := services.NewConversationService(&stubWorkflowRepo{
		customer: db.Customer{ID: customerID, TenantID: tenantA},
	})
	req := dto.ConversationMessageRequest{
		TenantID:   tenantA,
		CustomerID: customerID,
		Direction:  "in",
		Message:    "hello",
	}

	t.Run("cross-tenant customer returns ErrNotFound", func(t *testing.T) {
		_, err := svc.StoreMessage(context.Background(), req, &tenantB)
		require.ErrorIs(t, err, response.ErrNotFound)
	})
}
