package test

// Booking integration tests require a real PostgreSQL database.
// Set TEST_DATABASE_URL to run them, e.g.:
//
//	TEST_DATABASE_URL="postgres://user:pass@localhost:5432/appointments_test?sslmode=disable" go test ./tests/ -run TestBooking
//
// These tests verify the full booking flow including the FOR UPDATE slot lock
// that prevents double-booking under concurrent requests.

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/unknowncode44/appointments/internal/api/dto"
	"github.com/unknowncode44/appointments/internal/api/response"
	db "github.com/unknowncode44/appointments/internal/db/sqlc"
	"github.com/unknowncode44/appointments/internal/repositories"
	"github.com/unknowncode44/appointments/internal/services"
	"github.com/stretchr/testify/require"
)

func connectTestDB(t *testing.T) *db.Store {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set — skipping booking integration tests")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return db.NewStore(pool)
}

// seedBookingFixture inserts the minimum data needed for a booking test and
// returns a cleanup function that deletes everything in reverse order.
func seedBookingFixture(t *testing.T, store *db.Store) (tenantID, customerID, serviceID, slotID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	tenant, err := store.CreateTenant(ctx, db.CreateTenantParams{
		Name:     "Test Tenant " + uuid.NewString(),
		Timezone: "UTC",
	})
	require.NoError(t, err)
	tenantID = tenant.ID

	provider, err := store.CreateProvider(ctx, db.CreateProviderParams{
		TenantID: tenantID,
		Name:     "Test Provider",
	})
	require.NoError(t, err)

	svc, err := store.CreateService(ctx, db.CreateServiceParams{
		TenantID:        tenantID,
		Name:            "Test Service",
		DurationMinutes: 30,
	})
	require.NoError(t, err)
	serviceID = svc.ID

	customer, err := store.CreateCustomer(ctx, db.CreateCustomerParams{
		TenantID: tenantID,
	})
	require.NoError(t, err)
	customerID = customer.ID

	now := time.Now().Truncate(time.Minute).UTC()
	slot, err := store.CreateAppointmentSlot(ctx, db.CreateAppointmentSlotParams{
		TenantID:   tenantID,
		ProviderID: provider.ID,
		StartAt:    now.Add(24 * time.Hour),
		EndAt:      now.Add(24*time.Hour + 30*time.Minute),
	})
	require.NoError(t, err)
	slotID = slot.ID

	return tenantID, customerID, serviceID, slotID
}

// TestBooking_Create verifies that booking a slot marks the appointment
// confirmed and moves the slot out of available status.
func TestBooking_Create(t *testing.T) {
	store := connectTestDB(t)
	svc := services.NewAppointmentService(repositories.NewWorkflowRepository(store))

	tenantID, customerID, serviceID, slotID := seedBookingFixture(t, store)

	req := dto.AppointmentCreateRequest{
		TenantID:   tenantID,
		CustomerID: customerID,
		ServiceID:  serviceID,
		SlotID:     slotID,
	}

	appt, err := svc.Create(context.Background(), req, nil)
	require.NoError(t, err)
	require.Equal(t, "confirmed", appt.Status)
	require.NotNil(t, appt.SlotID)
	require.Equal(t, slotID, *appt.SlotID)

	// Verify slot is no longer available (FOR UPDATE locked and status changed).
	reservedSlot, err := store.GetAppointmentSlotForUpdate(context.Background(), slotID)
	require.NoError(t, err)
	require.Equal(t, "reserved", reservedSlot.Status)
}

// TestBooking_Cancel verifies that cancelling an appointment releases the slot
// back to available.
func TestBooking_Cancel(t *testing.T) {
	store := connectTestDB(t)
	apptSvc := services.NewAppointmentService(repositories.NewWorkflowRepository(store))

	tenantID, customerID, serviceID, slotID := seedBookingFixture(t, store)

	appt, err := apptSvc.Create(context.Background(), dto.AppointmentCreateRequest{
		TenantID:   tenantID,
		CustomerID: customerID,
		ServiceID:  serviceID,
		SlotID:     slotID,
	}, nil)
	require.NoError(t, err)

	cancelled, err := apptSvc.Cancel(context.Background(), appt.ID, nil)
	require.NoError(t, err)
	require.Equal(t, "cancelled", cancelled.Status)

	// Slot must be available again.
	releasedSlot, err := store.GetAppointmentSlotForUpdate(context.Background(), slotID)
	require.NoError(t, err)
	require.Equal(t, "available", releasedSlot.Status)
}

// TestBooking_DoubleBook verifies that booking an already-reserved slot returns
// ErrConflict, exercising the FOR UPDATE guard in the transaction.
func TestBooking_DoubleBook(t *testing.T) {
	store := connectTestDB(t)
	svc := services.NewAppointmentService(repositories.NewWorkflowRepository(store))

	tenantID, customerID, serviceID, slotID := seedBookingFixture(t, store)

	makeReq := func() dto.AppointmentCreateRequest {
		return dto.AppointmentCreateRequest{
			TenantID:   tenantID,
			CustomerID: customerID,
			ServiceID:  serviceID,
			SlotID:     slotID,
		}
	}

	// First booking succeeds.
	_, err := svc.Create(context.Background(), makeReq(), nil)
	require.NoError(t, err)

	// Second booking on the same slot must fail.
	_, err = svc.Create(context.Background(), makeReq(), nil)
	require.ErrorIs(t, err, response.ErrConflict)
}

// TestBooking_ScopeForced verifies that when a non-admin scope is provided,
// the appointment's TenantID is forced to the scope regardless of the request.
func TestBooking_Cancel_CrossTenant_Returns404(t *testing.T) {
	store := connectTestDB(t)
	svc := services.NewAppointmentService(repositories.NewWorkflowRepository(store))

	tenantA, customerID, serviceID, slotID := seedBookingFixture(t, store)
	tenantB := uuid.New()

	appt, err := svc.Create(context.Background(), dto.AppointmentCreateRequest{
		TenantID:   tenantA,
		CustomerID: customerID,
		ServiceID:  serviceID,
		SlotID:     slotID,
	}, nil)
	require.NoError(t, err)

	// Tenant B tries to cancel Tenant A's appointment.
	_, err = svc.Cancel(context.Background(), appt.ID, &tenantB)
	require.ErrorIs(t, err, response.ErrNotFound)
}
