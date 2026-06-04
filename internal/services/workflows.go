package services

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mousav1/ticket/internal/api/dto"
	"github.com/mousav1/ticket/internal/api/response"
	db "github.com/mousav1/ticket/internal/db/sqlc"
	"github.com/mousav1/ticket/internal/platform/pagination"
	"github.com/mousav1/ticket/internal/repositories"
)

type AppointmentService interface {
	Create(context.Context, dto.AppointmentCreateRequest) (dto.AppointmentResponse, error)
	List(context.Context, uuid.UUID, *uuid.UUID, *uuid.UUID, string, pagination.Page) (pagination.Response[dto.AppointmentResponse], error)
	Get(context.Context, uuid.UUID) (dto.AppointmentResponse, error)
	Update(context.Context, uuid.UUID, dto.AppointmentUpdateRequest) (dto.AppointmentResponse, error)
	Cancel(context.Context, uuid.UUID) (dto.AppointmentResponse, error)
}

type ConversationService interface {
	ListThreads(context.Context, uuid.UUID, pagination.Page) ([]dto.ConversationThreadResponse, error)
	GetThread(context.Context, uuid.UUID) (dto.ConversationThreadResponse, error)
	StoreMessage(context.Context, dto.ConversationMessageRequest) (dto.ConversationMessageResponse, error)
	ProcessInboundMessage(context.Context, dto.InboundMessageRequest) (dto.ConversationMessageResponse, error)
	ProcessEvolutionWebhook(context.Context, dto.EvolutionWebhookRequest, []byte) error
}

type appointmentService struct {
	repo repositories.WorkflowRepository
}
type conversationService struct {
	repo repositories.WorkflowRepository
}

func NewAppointmentService(repo repositories.WorkflowRepository) AppointmentService {
	return &appointmentService{repo: repo}
}

func NewConversationService(repo repositories.WorkflowRepository) ConversationService {
	return &conversationService{repo: repo}
}

func (s *appointmentService) Create(ctx context.Context, req dto.AppointmentCreateRequest) (dto.AppointmentResponse, error) {
	var created db.Appointment
	err := s.repo.Store().ExecTx(ctx, func(q *db.Queries) error {
		slot, err := q.GetAppointmentSlotForUpdate(ctx, req.SlotID)
		if err != nil {
			return err
		}
		if slot.Status != "available" || slot.TenantID != req.TenantID {
			return response.ErrConflict
		}
		appt, err := q.CreateAppointment(ctx, db.CreateAppointmentParams{
			TenantID: req.TenantID, CustomerID: req.CustomerID, ProviderID: slot.ProviderID, ServiceID: req.ServiceID,
			SlotID: uuid.NullUUID{UUID: slot.ID, Valid: true}, StartAt: slot.StartAt, EndAt: slot.EndAt,
			Status: "confirmed", Notes: repositories.Text(req.Notes),
		})
		if err != nil {
			return err
		}
		if _, err := q.ReserveAppointmentSlot(ctx, db.ReserveAppointmentSlotParams{ID: slot.ID, AppointmentID: appt.ID}); err != nil {
			return err
		}
		if _, err := q.CreateAppointmentEvent(ctx, db.CreateAppointmentEventParams{AppointmentID: appt.ID, EventType: "created", Payload: []byte(`{}`)}); err != nil {
			return err
		}
		if _, err := q.CreateAppointmentEvent(ctx, db.CreateAppointmentEventParams{AppointmentID: appt.ID, EventType: "confirmed", Payload: []byte(`{}`)}); err != nil {
			return err
		}
		created = appt
		return nil
	})
	return mapAppointment(created), err
}

func (s *appointmentService) List(ctx context.Context, tenantID uuid.UUID, providerID *uuid.UUID, customerID *uuid.UUID, status string, p pagination.Page) (pagination.Response[dto.AppointmentResponse], error) {
	items, total, err := s.repo.ListAppointments(ctx, db.ListAppointmentsParams{TenantID: tenantID, ProviderID: providerID, CustomerID: customerID, Status: status, Limit: p.PageSize, Offset: p.Offset})
	return paged(items, total, p, mapAppointment), err
}

func (s *appointmentService) Get(ctx context.Context, id uuid.UUID) (dto.AppointmentResponse, error) {
	item, err := s.repo.GetAppointment(ctx, id)
	return mapAppointment(item), err
}

func (s *appointmentService) Update(ctx context.Context, id uuid.UUID, req dto.AppointmentUpdateRequest) (dto.AppointmentResponse, error) {
	if req.Status != nil && *req.Status == "cancelled" {
		return s.Cancel(ctx, id)
	}
	var updated db.Appointment
	err := s.repo.Store().ExecTx(ctx, func(q *db.Queries) error {
		current, err := q.GetAppointment(ctx, id)
		if err != nil {
			return err
		}
		next := current
		if req.ServiceID != nil {
			next.ServiceID = *req.ServiceID
		}
		if req.Status != nil {
			next.Status = *req.Status
		}
		if req.Notes != nil {
			next.Notes = repositories.Text(req.Notes)
		}
		eventType := ""
		if req.SlotID != nil && (!current.SlotID.Valid || current.SlotID.UUID != *req.SlotID) {
			newSlot, err := q.GetAppointmentSlotForUpdate(ctx, *req.SlotID)
			if err != nil {
				return err
			}
			if newSlot.Status != "available" || newSlot.TenantID != current.TenantID {
				return response.ErrConflict
			}
			if current.SlotID.Valid {
				if _, err := q.ReleaseAppointmentSlot(ctx, current.SlotID.UUID); err != nil {
					return err
				}
			}
			next.ProviderID = newSlot.ProviderID
			next.SlotID = uuid.NullUUID{UUID: newSlot.ID, Valid: true}
			next.StartAt = newSlot.StartAt
			next.EndAt = newSlot.EndAt
			eventType = "rescheduled"
		}
		appt, err := q.UpdateAppointment(ctx, db.UpdateAppointmentParams{ID: id, ProviderID: next.ProviderID, ServiceID: next.ServiceID, SlotID: next.SlotID, StartAt: next.StartAt, EndAt: next.EndAt, Status: next.Status, Notes: next.Notes})
		if err != nil {
			return err
		}
		if req.SlotID != nil {
			if _, err := q.ReserveAppointmentSlot(ctx, db.ReserveAppointmentSlotParams{ID: *req.SlotID, AppointmentID: id}); err != nil {
				return err
			}
		}
		if eventType != "" {
			if _, err := q.CreateAppointmentEvent(ctx, db.CreateAppointmentEventParams{AppointmentID: id, EventType: eventType, Payload: []byte(`{}`)}); err != nil {
				return err
			}
		}
		updated = appt
		return nil
	})
	return mapAppointment(updated), err
}

func (s *appointmentService) Cancel(ctx context.Context, id uuid.UUID) (dto.AppointmentResponse, error) {
	var cancelled db.Appointment
	err := s.repo.Store().ExecTx(ctx, func(q *db.Queries) error {
		current, err := q.GetAppointment(ctx, id)
		if err != nil {
			return err
		}
		if current.SlotID.Valid {
			if _, err := q.ReleaseAppointmentSlot(ctx, current.SlotID.UUID); err != nil {
				return err
			}
		}
		appt, err := q.CancelAppointment(ctx, id)
		if err != nil {
			return err
		}
		if _, err := q.CreateAppointmentEvent(ctx, db.CreateAppointmentEventParams{AppointmentID: id, EventType: "cancelled", Payload: []byte(`{}`)}); err != nil {
			return err
		}
		cancelled = appt
		return nil
	})
	return mapAppointment(cancelled), err
}

func (s *conversationService) ListThreads(ctx context.Context, tenantID uuid.UUID, p pagination.Page) ([]dto.ConversationThreadResponse, error) {
	items, err := s.repo.ListThreads(ctx, db.ListConversationThreadsParams{TenantID: tenantID, Limit: p.PageSize, Offset: p.Offset})
	return mapSlice(items, mapThread), err
}

func (s *conversationService) GetThread(ctx context.Context, id uuid.UUID) (dto.ConversationThreadResponse, error) {
	thread, err := s.repo.GetThread(ctx, id)
	if err != nil {
		return dto.ConversationThreadResponse{}, err
	}
	messages, err := s.repo.ListMessages(ctx, id)
	if err != nil {
		return dto.ConversationThreadResponse{}, err
	}
	resp := mapThread(thread)
	resp.Messages = mapSlice(messages, mapMessage)
	return resp, nil
}

func (s *conversationService) StoreMessage(ctx context.Context, req dto.ConversationMessageRequest) (dto.ConversationMessageResponse, error) {
	var out db.ConversationMessage
	err := s.repo.Store().ExecTx(ctx, func(q *db.Queries) error {
		thread, err := q.GetConversationThreadByCustomer(ctx, db.CreateConversationThreadParams{TenantID: req.TenantID, CustomerID: req.CustomerID})
		if errors.Is(err, pgx.ErrNoRows) {
			thread, err = q.CreateConversationThread(ctx, db.CreateConversationThreadParams{TenantID: req.TenantID, CustomerID: req.CustomerID})
		}
		if err != nil {
			return err
		}
		msg, err := q.CreateConversationMessage(ctx, db.CreateConversationMessageParams{ThreadID: thread.ID, Direction: req.Direction, Message: req.Message, Metadata: req.Metadata})
		if err != nil {
			return err
		}
		_, err = q.UpsertConversationState(ctx, db.UpsertConversationStateParams{TenantID: req.TenantID, CustomerID: req.CustomerID, State: "message_received", Data: []byte(`{}`)})
		out = msg
		return err
	})
	return mapMessage(out), err
}

func (s *conversationService) ProcessInboundMessage(ctx context.Context, req dto.InboundMessageRequest) (dto.ConversationMessageResponse, error) {
	source := req.Source
	if source == "" {
		source = "n8n"
	}
	metadata := req.Metadata
	if len(metadata) == 0 {
		metadata, _ = json.Marshal(req)
	}
	msg, err := s.storeInboundMessage(ctx, inboundMessage{
		ChannelExternalID:  req.TenantChannelExternalID,
		ExternalCustomerID: req.ExternalCustomerID,
		Message:            req.Message,
		Source:             source,
		Metadata:           metadata,
		LogPayload:         metadata,
		State:              "inbound_message_received",
	})
	return mapMessage(msg), err
}

func (s *conversationService) ProcessEvolutionWebhook(ctx context.Context, req dto.EvolutionWebhookRequest, raw []byte) error {
	externalID := req.ExternalID
	if externalID == "" {
		externalID = req.Instance
	}
	externalCustomer := req.From
	if externalCustomer == "" {
		externalCustomer = req.Phone
	}
	if externalID == "" || externalCustomer == "" {
		return response.ErrInvalidInput
	}
	message := req.Message
	if message == "" && len(req.Data) > 0 {
		message = string(req.Data)
	}
	metadata, _ := json.Marshal(req)
	_, err := s.storeInboundMessage(ctx, inboundMessage{
		ChannelExternalID:  externalID,
		ExternalCustomerID: externalCustomer,
		Message:            message,
		Source:             "evolution",
		Metadata:           metadata,
		LogPayload:         raw,
		State:              "webhook_received",
	})
	return err
}

type inboundMessage struct {
	ChannelExternalID  string
	ExternalCustomerID string
	Message            string
	Source             string
	Metadata           []byte
	LogPayload         []byte
	State              string
}

func (s *conversationService) storeInboundMessage(ctx context.Context, req inboundMessage) (db.ConversationMessage, error) {
	var out db.ConversationMessage
	if req.ChannelExternalID == "" || req.ExternalCustomerID == "" || req.Message == "" {
		return out, response.ErrInvalidInput
	}
	if len(req.Metadata) == 0 {
		req.Metadata = []byte(`{}`)
	}
	if len(req.LogPayload) == 0 {
		req.LogPayload = req.Metadata
	}
	err := s.repo.Store().ExecTx(ctx, func(q *db.Queries) error {
		channel, err := q.GetTenantChannelByExternalID(ctx, req.ChannelExternalID)
		if err != nil {
			return err
		}
		if _, err := q.CreateWebhookLog(ctx, db.CreateWebhookLogParams{TenantID: uuid.NullUUID{UUID: channel.TenantID, Valid: true}, Source: req.Source, Payload: req.LogPayload}); err != nil {
			return err
		}
		customer, err := q.GetCustomerByChannel(ctx, db.GetCustomerByChannelParams{TenantChannelID: channel.ID, ExternalIdentifier: req.ExternalCustomerID})
		if errors.Is(err, pgx.ErrNoRows) {
			customer, err = q.CreateCustomer(ctx, db.CreateCustomerParams{TenantID: channel.TenantID})
			if err != nil {
				return err
			}
			if _, err := q.CreateCustomerChannel(ctx, db.CreateCustomerChannelParams{CustomerID: customer.ID, TenantChannelID: channel.ID, ExternalIdentifier: req.ExternalCustomerID}); err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
		thread, err := q.GetConversationThreadByCustomer(ctx, db.CreateConversationThreadParams{TenantID: channel.TenantID, CustomerID: customer.ID})
		if errors.Is(err, pgx.ErrNoRows) {
			thread, err = q.CreateConversationThread(ctx, db.CreateConversationThreadParams{TenantID: channel.TenantID, CustomerID: customer.ID})
		}
		if err != nil {
			return err
		}
		msg, err := q.CreateConversationMessage(ctx, db.CreateConversationMessageParams{ThreadID: thread.ID, Direction: "in", Message: req.Message, Metadata: req.Metadata})
		if err != nil {
			return err
		}
		_, err = q.UpsertConversationState(ctx, db.UpsertConversationStateParams{TenantID: channel.TenantID, CustomerID: customer.ID, State: req.State, Data: req.Metadata})
		out = msg
		return err
	})
	return out, err
}
