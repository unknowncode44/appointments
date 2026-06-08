package db

import (
	"context"

	"github.com/google/uuid"
)

// GetWhatsappChannelByTenant returns any WhatsApp channel record for the tenant,
// regardless of active status. Used to prevent duplicate instance creation.
func (q *Queries) GetWhatsappChannelByTenant(ctx context.Context, tenantID uuid.UUID) (TenantChannel, error) {
	var i TenantChannel
	err := q.db.QueryRow(ctx,
		"SELECT "+tenantChannelCols+" FROM tenant_channels WHERE tenant_id = $1 AND channel_type = 'whatsapp' LIMIT 1",
		tenantID,
	).Scan(&i.ID, &i.TenantID, &i.ChannelType, &i.ExternalID, &i.ExternalKey, &i.Active, &i.CreatedAt)
	return i, err
}

// DeactivateWhatsappChannelByTenant sets the tenant's WhatsApp channel to inactive.
func (q *Queries) DeactivateWhatsappChannelByTenant(ctx context.Context, tenantID uuid.UUID) error {
	_, err := q.db.Exec(ctx,
		"UPDATE tenant_channels SET active = false WHERE tenant_id = $1 AND channel_type = 'whatsapp'",
		tenantID,
	)
	return err
}

// DeleteWhatsappChannelByTenant hard-deletes the tenant's WhatsApp channel record.
func (q *Queries) DeleteWhatsappChannelByTenant(ctx context.Context, tenantID uuid.UUID) error {
	_, err := q.db.Exec(ctx,
		"DELETE FROM tenant_channels WHERE tenant_id = $1 AND channel_type = 'whatsapp'",
		tenantID,
	)
	return err
}
