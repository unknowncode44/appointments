package test

// WhatsApp bot integration tests require a real PostgreSQL database.
// Set TEST_DATABASE_URL to run them (see booking_test.go). They cover webhook
// idempotency and the internal outbound sender.

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/unknowncode44/appointments/internal/api/dto"
	db "github.com/unknowncode44/appointments/internal/db/sqlc"
	"github.com/unknowncode44/appointments/internal/repositories"
	"github.com/unknowncode44/appointments/internal/services"
)

// mockSender records SendText calls and never fails.
type mockSender struct {
	calls []sentMessage
}

type sentMessage struct{ instance, apiKey, to, text string }

func (m *mockSender) SendText(_ context.Context, instance, apiKey, to, text string) error {
	m.calls = append(m.calls, sentMessage{instance, apiKey, to, text})
	return nil
}

func seedChannel(t *testing.T, store *db.Store, externalID string) (tenantID uuid.UUID, channel db.TenantChannel) {
	t.Helper()
	ctx := context.Background()
	tenant, err := store.CreateTenant(ctx, db.CreateTenantParams{Name: "WA Tenant " + uuid.NewString(), Timezone: "UTC"})
	require.NoError(t, err)
	channel, err = store.CreateTenantChannel(ctx, db.CreateTenantChannelParams{
		TenantID:    tenant.ID,
		ChannelType: "whatsapp",
		ExternalID:  externalID,
		ExternalKey: pgtype.Text{String: "instance-key", Valid: true},
	})
	require.NoError(t, err)
	return tenant.ID, channel
}

func evolutionPayload(instance, messageID, from, text string) dto.EvolutionWebhookRequest {
	var req dto.EvolutionWebhookRequest
	req.Instance = instance
	req.Data.Key.ID = messageID
	req.Data.Key.RemoteJid = from
	req.Data.Message.Conversation = text
	return req
}

// TestWebhook_Idempotency verifies the same Evolution message id is processed
// once; a redelivery is reported idempotent and stores no second inbound message.
func TestWebhook_Idempotency(t *testing.T) {
	store := connectTestDB(t)
	sender := &mockSender{}
	svc := services.NewConversationService(repositories.NewWorkflowRepository(store), sender)

	externalID := "turnobot_" + uuid.NewString()
	_, _ = seedChannel(t, store, externalID)

	req := evolutionPayload(externalID, "msg-"+uuid.NewString(), "5491100000000@s.whatsapp.net", "hola")
	raw := []byte(`{}`)

	first, err := svc.ProcessEvolutionWebhook(context.Background(), req, raw)
	require.NoError(t, err)
	require.True(t, first.Processed)
	require.False(t, first.Idempotent)
	require.NotNil(t, first.CustomerID)

	second, err := svc.ProcessEvolutionWebhook(context.Background(), req, raw)
	require.NoError(t, err)
	require.False(t, second.Processed)
	require.True(t, second.Idempotent)

	// Only the first delivery stored an inbound conversation message.
	thread, err := store.GetConversationThreadByCustomer(context.Background(), db.CreateConversationThreadParams{
		TenantID:   *first.TenantID,
		CustomerID: *first.CustomerID,
	})
	require.NoError(t, err)
	messages, err := store.ListConversationMessages(context.Background(), thread.ID)
	require.NoError(t, err)
	inbound := 0
	for _, m := range messages {
		if m.Direction == "in" {
			inbound++
		}
	}
	require.Equal(t, 1, inbound)
}

// TestSendWhatsAppText verifies the sender persists an outbound row plus an
// outbound conversation message and calls the Evolution client.
func TestSendWhatsAppText(t *testing.T) {
	store := connectTestDB(t)
	sender := &mockSender{}
	svc := services.NewConversationService(repositories.NewWorkflowRepository(store), sender)

	externalID := "turnobot_" + uuid.NewString()
	tenantID, channel := seedChannel(t, store, externalID)

	customer, err := store.CreateCustomer(context.Background(), db.CreateCustomerParams{TenantID: tenantID})
	require.NoError(t, err)

	to := "5491100000000@s.whatsapp.net"
	err = svc.SendWhatsAppText(context.Background(), channel, customer.ID, to, "tu turno está confirmado")
	require.NoError(t, err)

	// Evolution client was called with the per-instance id and key.
	require.Len(t, sender.calls, 1)
	require.Equal(t, channel.ExternalID, sender.calls[0].instance)
	require.Equal(t, channel.ExternalKey.String, sender.calls[0].apiKey)
	require.Equal(t, to, sender.calls[0].to)

	// An outbound conversation message (direction "out") was recorded.
	thread, err := store.GetConversationThreadByCustomer(context.Background(), db.CreateConversationThreadParams{TenantID: tenantID, CustomerID: customer.ID})
	require.NoError(t, err)
	messages, err := store.ListConversationMessages(context.Background(), thread.ID)
	require.NoError(t, err)
	out := 0
	for _, m := range messages {
		if m.Direction == "out" {
			out++
		}
	}
	require.Equal(t, 1, out)

	// An outbound_messages row was persisted and marked sent.
	pool, err := pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))
	require.NoError(t, err)
	defer pool.Close()
	var count int
	err = pool.QueryRow(context.Background(),
		"SELECT count(*) FROM outbound_messages WHERE tenant_id = $1 AND customer_id = $2 AND status = 'sent'",
		tenantID, customer.ID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
