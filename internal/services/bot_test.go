package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/unknowncode44/appointments/internal/api/response"
	db "github.com/unknowncode44/appointments/internal/db/sqlc"
)

// stubBotRepo is an in-memory botRepo for FSM unit tests.
type stubBotRepo struct {
	tenant    db.Tenant
	services  []db.Service
	providers []db.Provider
	slots     []db.AppointmentSlot

	bookErr     error
	booked      *bookSlotParams
	slotsTenant uuid.UUID // tenant id seen by ListAvailableSlots
}

func (r *stubBotRepo) GetTenant(_ context.Context, _ uuid.UUID) (db.Tenant, error) {
	return r.tenant, nil
}
func (r *stubBotRepo) ListServices(_ context.Context, _ db.ListServicesParams) ([]db.Service, error) {
	return r.services, nil
}
func (r *stubBotRepo) ListProviders(_ context.Context, _ db.ListProvidersParams) ([]db.Provider, error) {
	return r.providers, nil
}
func (r *stubBotRepo) ListAvailableSlots(_ context.Context, arg db.ListAvailableSlotsParams) ([]db.AppointmentSlot, error) {
	r.slotsTenant = arg.TenantID
	return r.slots, nil
}
func (r *stubBotRepo) BookSlot(_ context.Context, p bookSlotParams) (db.Appointment, error) {
	pc := p
	r.booked = &pc
	if r.bookErr != nil {
		return db.Appointment{}, r.bookErr
	}
	return db.Appointment{ID: uuid.New(), Status: "confirmed"}, nil
}

func fixedBot(repo *stubBotRepo) *bot {
	b := newBot(repo)
	b.now = func() time.Time { return time.Date(2026, 6, 29, 9, 0, 0, 0, time.UTC) }
	return b
}

func TestBot_HappyPath_SingleProvider(t *testing.T) {
	ctx := context.Background()
	svcID, provID, slotID := uuid.New(), uuid.New(), uuid.New()
	tenantID, customerID := uuid.New(), uuid.New()

	repo := &stubBotRepo{
		tenant:    db.Tenant{ID: tenantID, Timezone: "UTC"},
		services:  []db.Service{{ID: svcID, Name: "Corte"}},
		providers: []db.Provider{{ID: provID, Name: "Ana"}},
		slots: []db.AppointmentSlot{{
			ID:      slotID,
			StartAt: time.Date(2026, 6, 29, 14, 0, 0, 0, time.UTC),
			EndAt:   time.Date(2026, 6, 29, 14, 30, 0, 0, time.UTC),
		}},
	}
	b := fixedBot(repo)
	in := botInput{TenantID: tenantID, CustomerID: customerID}

	// greeting
	out, err := b.Handle(ctx, withText(in, "", botData{}, "hola"))
	require.NoError(t, err)
	require.Equal(t, stepChooseService, out.Step)
	require.Contains(t, out.Reply, "Corte")

	// choose service -> single provider auto-selected -> dates
	out, err = b.Handle(ctx, withText(in, out.Step, out.Data, "1"))
	require.NoError(t, err)
	require.Equal(t, stepChooseDate, out.Step)
	require.Equal(t, svcID.String(), out.Data.ServiceID)
	require.Equal(t, provID.String(), out.Data.ProviderID)

	// choose date "1" (today) -> slots
	out, err = b.Handle(ctx, withText(in, out.Step, out.Data, "1"))
	require.NoError(t, err)
	require.Equal(t, stepChooseSlot, out.Step)
	require.Equal(t, "2026-06-29", out.Data.Date)
	require.Contains(t, out.Reply, "14:00")

	// choose slot "1" -> confirm
	out, err = b.Handle(ctx, withText(in, out.Step, out.Data, "1"))
	require.NoError(t, err)
	require.Equal(t, stepConfirm, out.Step)
	require.Equal(t, slotID.String(), out.Data.SlotID)

	// confirm "1" -> booked via BookSlot
	out, err = b.Handle(ctx, withText(in, out.Step, out.Data, "1"))
	require.NoError(t, err)
	require.Equal(t, stepBooked, out.Step)
	require.Contains(t, out.Reply, "confirmado")
	require.NotNil(t, repo.booked)
	require.Equal(t, tenantID, repo.booked.TenantID)
	require.Equal(t, customerID, repo.booked.CustomerID)
	require.Equal(t, svcID, repo.booked.ServiceID)
	require.Equal(t, slotID, repo.booked.SlotID)
}

func TestBot_MultiProvider_OffersAnyOption(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	repo := &stubBotRepo{
		tenant:    db.Tenant{ID: tenantID, Timezone: "UTC"},
		services:  []db.Service{{ID: uuid.New(), Name: "Corte"}},
		providers: []db.Provider{{ID: uuid.New(), Name: "Ana"}, {ID: uuid.New(), Name: "Beto"}},
	}
	b := fixedBot(repo)
	in := botInput{TenantID: tenantID, CustomerID: uuid.New()}

	out, _ := b.Handle(ctx, withText(in, stepChooseService, botData{
		Offered: []offeredItem{{ID: repo.services[0].ID.String(), Label: "Corte"}},
	}, "1"))
	require.Equal(t, stepChooseProvider, out.Step)
	require.Contains(t, out.Reply, "Cualquier profesional")

	// pick "any" (option 3) -> dates with no provider filter
	out, _ = b.Handle(ctx, withText(in, out.Step, out.Data, "3"))
	require.Equal(t, stepChooseDate, out.Step)
	require.Equal(t, "", out.Data.ProviderID)
}

func TestBot_InvalidChoiceReprompts(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	repo := &stubBotRepo{
		tenant:   db.Tenant{ID: tenantID, Timezone: "UTC"},
		services: []db.Service{{ID: uuid.New(), Name: "Corte"}},
	}
	b := fixedBot(repo)
	in := botInput{TenantID: tenantID, CustomerID: uuid.New()}

	data := botData{Offered: []offeredItem{{ID: repo.services[0].ID.String(), Label: "Corte"}}}
	out, err := b.Handle(ctx, withText(in, stepChooseService, data, "no-soy-un-numero"))
	require.NoError(t, err)
	require.Equal(t, stepChooseService, out.Step) // stays put
	require.Contains(t, out.Reply, "No entendí")
}

func TestBot_ConfirmConflict_ReoffersSlots(t *testing.T) {
	ctx := context.Background()
	svcID, slotID := uuid.New(), uuid.New()
	tenantID, customerID := uuid.New(), uuid.New()

	repo := &stubBotRepo{
		tenant:  db.Tenant{ID: tenantID, Timezone: "UTC"},
		bookErr: response.ErrConflict,
		slots: []db.AppointmentSlot{{
			ID:      uuid.New(),
			StartAt: time.Date(2026, 6, 29, 15, 0, 0, 0, time.UTC),
			EndAt:   time.Date(2026, 6, 29, 15, 30, 0, 0, time.UTC),
		}},
	}
	b := fixedBot(repo)

	in := withText(botInput{TenantID: tenantID, CustomerID: customerID}, stepConfirm, botData{
		ServiceID:   svcID.String(),
		ServiceName: "Corte",
		SlotID:      slotID.String(),
		Date:        "2026-06-29",
	}, "1")

	out, err := b.Handle(ctx, in)
	require.NoError(t, err)
	require.Equal(t, stepChooseSlot, out.Step)
	require.Contains(t, out.Reply, "se ocupó")
	require.Contains(t, out.Reply, "15:00")
	// Tenant scoping: the re-offer query used the input tenant, never one from data.
	require.Equal(t, tenantID, repo.slotsTenant)
}

func TestBot_BookingIsTenantScoped(t *testing.T) {
	ctx := context.Background()
	svcID, slotID := uuid.New(), uuid.New()
	tenantID, customerID := uuid.New(), uuid.New()
	repo := &stubBotRepo{tenant: db.Tenant{ID: tenantID, Timezone: "UTC"}}
	b := fixedBot(repo)

	in := withText(botInput{TenantID: tenantID, CustomerID: customerID}, stepConfirm, botData{
		ServiceID: svcID.String(),
		SlotID:    slotID.String(),
		Date:      "2026-06-29",
	}, "si")

	_, err := b.Handle(ctx, in)
	require.NoError(t, err)
	require.NotNil(t, repo.booked)
	// The booking always uses the instance-derived tenant and customer, not input.
	require.Equal(t, tenantID, repo.booked.TenantID)
	require.Equal(t, customerID, repo.booked.CustomerID)
}

func withText(in botInput, step string, data botData, text string) botInput {
	in.Step = step
	in.Data = data
	in.Text = text
	return in
}
