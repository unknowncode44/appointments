package repositories

import (
	"context"

	"github.com/google/uuid"
	db "github.com/unknowncode44/appointments/internal/db/sqlc"
)

type WorkflowRepository interface {
	Store() *db.Store
	ListAppointments(context.Context, db.ListAppointmentsParams) ([]db.Appointment, int64, error)
	GetAppointment(context.Context, uuid.UUID) (db.Appointment, error)
	ListThreads(context.Context, db.ListConversationThreadsParams) ([]db.ConversationThread, error)
	GetThread(context.Context, uuid.UUID) (db.ConversationThread, error)
	ListMessages(context.Context, uuid.UUID) ([]db.ConversationMessage, error)
	GetTenantChannelByExternalID(context.Context, string) (db.TenantChannel, error)
}

type workflowRepository struct{ store *db.Store }

func NewWorkflowRepository(store *db.Store) WorkflowRepository {
	return &workflowRepository{store: store}
}

func (r *workflowRepository) Store() *db.Store { return r.store }

func (r *workflowRepository) ListAppointments(ctx context.Context, arg db.ListAppointmentsParams) ([]db.Appointment, int64, error) {
	items, err := r.store.ListAppointments(ctx, arg)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.store.CountAppointments(ctx, arg)
	return items, total, err
}
func (r *workflowRepository) GetAppointment(ctx context.Context, id uuid.UUID) (db.Appointment, error) {
	return r.store.GetAppointment(ctx, id)
}
func (r *workflowRepository) ListThreads(ctx context.Context, arg db.ListConversationThreadsParams) ([]db.ConversationThread, error) {
	return r.store.ListConversationThreads(ctx, arg)
}
func (r *workflowRepository) GetThread(ctx context.Context, id uuid.UUID) (db.ConversationThread, error) {
	return r.store.GetConversationThread(ctx, id)
}
func (r *workflowRepository) ListMessages(ctx context.Context, threadID uuid.UUID) ([]db.ConversationMessage, error) {
	return r.store.ListConversationMessages(ctx, threadID)
}
func (r *workflowRepository) GetTenantChannelByExternalID(ctx context.Context, externalID string) (db.TenantChannel, error) {
	return r.store.GetTenantChannelByExternalID(ctx, externalID)
}
