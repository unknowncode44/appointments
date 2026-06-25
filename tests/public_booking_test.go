package test

// Public booking integration tests require a real PostgreSQL database.
// Set TEST_DATABASE_URL to run them (see booking_test.go). They exercise the
// public, no-auth booking flow: slug resolution, the confirmed-booking path,
// customer-by-phone idempotency, cross-tenant safety and the FOR UPDATE guard.

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/unknowncode44/appointments/internal/api/dto"
	"github.com/unknowncode44/appointments/internal/api/response"
	db "github.com/unknowncode44/appointments/internal/db/sqlc"
	"github.com/unknowncode44/appointments/internal/repositories"
	"github.com/unknowncode44/appointments/internal/services"
)

func text(v string) pgtype.Text { return pgtype.Text{String: v, Valid: true} }

// seedPublicTenant creates a tenant with a unique slug plus a provider and an
// active service. The slug is returned for use against the public service.
func seedPublicTenant(t *testing.T, store *db.Store) (tenantID, providerID, serviceID uuid.UUID, slug string) {
	t.Helper()
	ctx := context.Background()

	slug = "shop-" + uuid.NewString()
	tenant, err := store.CreateTenant(ctx, db.CreateTenantParams{
		Name:     "Public Tenant " + uuid.NewString(),
		Timezone: "UTC",
		Slug:     text(slug),
	})
	require.NoError(t, err)
	tenantID = tenant.ID

	provider, err := store.CreateProvider(ctx, db.CreateProviderParams{TenantID: tenantID, Name: "Public Provider"})
	require.NoError(t, err)
	providerID = provider.ID

	svc, err := store.CreateService(ctx, db.CreateServiceParams{TenantID: tenantID, Name: "Cut", DurationMinutes: 30})
	require.NoError(t, err)
	serviceID = svc.ID

	return tenantID, providerID, serviceID, slug
}

// createSlot inserts a single available slot for a provider, offset hours into
// the future so each call yields a distinct time.
func createSlot(t *testing.T, store *db.Store, tenantID, providerID uuid.UUID, hoursAhead int) uuid.UUID {
	t.Helper()
	now := time.Now().Truncate(time.Minute).UTC()
	start := now.Add(time.Duration(hoursAhead) * time.Hour)
	slot, err := store.CreateAppointmentSlot(context.Background(), db.CreateAppointmentSlotParams{
		TenantID:   tenantID,
		ProviderID: providerID,
		StartAt:    start,
		EndAt:      start.Add(30 * time.Minute),
	})
	require.NoError(t, err)
	return slot.ID
}

func newPublicService(store *db.Store) services.PublicService {
	return services.NewPublicService(repositories.NewPublicRepository(store))
}

// TestPublic_ResolveTenant_UnknownSlug verifies an unknown slug is not found.
func TestPublic_ResolveTenant_UnknownSlug(t *testing.T) {
	store := connectTestDB(t)
	svc := newPublicService(store)

	_, err := svc.ResolveTenant(context.Background(), "does-not-exist-"+uuid.NewString())
	require.ErrorIs(t, err, pgx.ErrNoRows)
}

// TestPublic_ResolveTenant_InactiveSlug verifies an inactive tenant's slug is
// not resolvable (GetTenantBySlug filters active = true).
func TestPublic_ResolveTenant_InactiveSlug(t *testing.T) {
	store := connectTestDB(t)
	svc := newPublicService(store)

	tenantID, _, _, slug := seedPublicTenant(t, store)
	_, err := store.DeactivateTenant(context.Background(), tenantID)
	require.NoError(t, err)

	_, err = svc.ResolveTenant(context.Background(), slug)
	require.ErrorIs(t, err, pgx.ErrNoRows)
}

// TestPublic_Booking_HappyPath verifies a confirmed appointment is created, the
// slot becomes reserved and the customer is created from the booking payload.
func TestPublic_Booking_HappyPath(t *testing.T) {
	store := connectTestDB(t)
	svc := newPublicService(store)

	tenantID, providerID, serviceID, slug := seedPublicTenant(t, store)
	slotID := createSlot(t, store, tenantID, providerID, 24)

	resolved, err := svc.ResolveTenant(context.Background(), slug)
	require.NoError(t, err)
	require.Equal(t, tenantID, resolved.ID)

	appt, err := svc.CreateBooking(context.Background(), resolved.ID, dto.PublicBookingRequest{
		FirstName: "Ada",
		Phone:     "+5491100000001",
		ServiceID: serviceID,
		SlotID:    slotID,
	})
	require.NoError(t, err)
	require.Equal(t, "confirmed", appt.Status)
	require.Equal(t, providerID, appt.ProviderID)

	reserved, err := store.GetAppointmentSlotForUpdate(context.Background(), slotID)
	require.NoError(t, err)
	require.Equal(t, "reserved", reserved.Status)
}

// TestPublic_Booking_CustomerIdempotency verifies two bookings with the same
// phone reuse one customer.
func TestPublic_Booking_CustomerIdempotency(t *testing.T) {
	store := connectTestDB(t)
	svc := newPublicService(store)

	tenantID, providerID, serviceID, _ := seedPublicTenant(t, store)
	slot1 := createSlot(t, store, tenantID, providerID, 24)
	slot2 := createSlot(t, store, tenantID, providerID, 25)

	phone := "+5491100000002"
	mk := func(slotID uuid.UUID) dto.PublicBookingRequest {
		return dto.PublicBookingRequest{FirstName: "Bob", Phone: phone, ServiceID: serviceID, SlotID: slotID}
	}

	a1, err := svc.CreateBooking(context.Background(), tenantID, mk(slot1))
	require.NoError(t, err)
	a2, err := svc.CreateBooking(context.Background(), tenantID, mk(slot2))
	require.NoError(t, err)

	c1, err := store.GetAppointment(context.Background(), a1.ID)
	require.NoError(t, err)
	c2, err := store.GetAppointment(context.Background(), a2.ID)
	require.NoError(t, err)
	require.Equal(t, c1.CustomerID, c2.CustomerID, "same phone must reuse one customer")
}

// TestPublic_Booking_CrossTenantSlot verifies that booking a slot belonging to a
// different tenant than the slug is rejected (404, never succeeds).
func TestPublic_Booking_CrossTenantSlot(t *testing.T) {
	store := connectTestDB(t)
	svc := newPublicService(store)

	tenantA, _, serviceA, _ := seedPublicTenant(t, store)
	tenantB, providerB, _, _ := seedPublicTenant(t, store)
	foreignSlot := createSlot(t, store, tenantB, providerB, 24)

	_, err := svc.CreateBooking(context.Background(), tenantA, dto.PublicBookingRequest{
		FirstName: "Eve",
		Phone:     "+5491100000003",
		ServiceID: serviceA, // valid service in A
		SlotID:    foreignSlot,
	})
	require.ErrorIs(t, err, response.ErrNotFound)
}

// TestPublic_Booking_DoubleBook verifies that two concurrent bookings of the
// same slot result in exactly one success (the FOR UPDATE guard).
func TestPublic_Booking_DoubleBook(t *testing.T) {
	store := connectTestDB(t)
	svc := newPublicService(store)

	tenantID, providerID, serviceID, _ := seedPublicTenant(t, store)
	slotID := createSlot(t, store, tenantID, providerID, 24)

	mk := func(phone string) dto.PublicBookingRequest {
		return dto.PublicBookingRequest{FirstName: "Race", Phone: phone, ServiceID: serviceID, SlotID: slotID}
	}

	var wg sync.WaitGroup
	errs := make([]error, 2)
	phones := []string{"+5491100000004", "+5491100000005"}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, errs[idx] = svc.CreateBooking(context.Background(), tenantID, mk(phones[idx]))
		}(i)
	}
	wg.Wait()

	successes, conflicts := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, response.ErrConflict):
			conflicts++
		default:
			t.Fatalf("unexpected booking error: %v", err)
		}
	}
	require.Equal(t, 1, successes, "exactly one booking should succeed")
	require.Equal(t, 1, conflicts, "the other should conflict")
}
