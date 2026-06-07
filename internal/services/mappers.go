package services

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/mousav1/ticket/internal/api/dto"
	db "github.com/mousav1/ticket/internal/db/sqlc"
	"github.com/mousav1/ticket/internal/repositories"
)

func mapTenant(v db.Tenant) dto.TenantResponse {
	return dto.TenantResponse(v)
}

func mapProvider(v db.Provider) dto.ProviderResponse {
	return dto.ProviderResponse(v)
}

func mapService(v db.Service) dto.ServiceResponse {
	return dto.ServiceResponse(v)
}

func mapCustomer(v db.Customer) dto.CustomerResponse {
	return dto.CustomerResponse{
		ID:        v.ID,
		TenantID:  v.TenantID,
		FirstName: repositories.TextPtr(v.FirstName),
		LastName:  repositories.TextPtr(v.LastName),
		Notes:     repositories.TextPtr(v.Notes),
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}

func mapTenantChannel(v db.TenantChannel) dto.TenantChannelResponse {
	return dto.TenantChannelResponse{
		ID:          v.ID,
		TenantID:    v.TenantID,
		ChannelType: v.ChannelType,
		ExternalID:  v.ExternalID,
		ExternalKey: repositories.TextPtr(v.ExternalKey),
		Active:      v.Active,
		CreatedAt:   v.CreatedAt,
	}
}

func mapAvailability(v db.ProviderAvailability) dto.AvailabilityResponse {
	return dto.AvailabilityResponse{
		ID:         v.ID,
		ProviderID: v.ProviderID,
		Weekday:    v.Weekday,
		StartTime:  v.StartTime.Format("15:04:05"),
		EndTime:    v.EndTime.Format("15:04:05"),
	}
}

func mapException(v db.ProviderException) dto.ExceptionResponse {
	return dto.ExceptionResponse{
		ID:         v.ID,
		ProviderID: v.ProviderID,
		StartAt:    v.StartAt,
		EndAt:      v.EndAt,
		Reason:     repositories.TextPtr(v.Reason),
		CreatedAt:  v.CreatedAt,
	}
}

func mapSlot(v db.AppointmentSlot) dto.SlotResponse {
	var appointmentID *uuid.UUID
	if v.AppointmentID.Valid {
		appointmentID = &v.AppointmentID.UUID
	}
	return dto.SlotResponse{
		ID:            v.ID,
		TenantID:      v.TenantID,
		ProviderID:    v.ProviderID,
		StartAt:       v.StartAt,
		EndAt:         v.EndAt,
		Status:        v.Status,
		AppointmentID: appointmentID,
		CreatedAt:     v.CreatedAt,
	}
}

func mapAppointment(v db.Appointment) dto.AppointmentResponse {
	var slotID *uuid.UUID
	if v.SlotID.Valid {
		slotID = &v.SlotID.UUID
	}
	return dto.AppointmentResponse{
		ID:         v.ID,
		TenantID:   v.TenantID,
		CustomerID: v.CustomerID,
		ProviderID: v.ProviderID,
		ServiceID:  v.ServiceID,
		SlotID:     slotID,
		StartAt:    v.StartAt,
		EndAt:      v.EndAt,
		Status:     v.Status,
		Notes:      repositories.TextPtr(v.Notes),
		CreatedAt:  v.CreatedAt,
		UpdatedAt:  v.UpdatedAt,
	}
}

func mapThread(v db.ConversationThread) dto.ConversationThreadResponse {
	return dto.ConversationThreadResponse{
		ID:         v.ID,
		TenantID:   v.TenantID,
		CustomerID: v.CustomerID,
		CreatedAt:  v.CreatedAt,
	}
}

func mapMessage(v db.ConversationMessage) dto.ConversationMessageResponse {
	return dto.ConversationMessageResponse{
		ID:        v.ID,
		ThreadID:  v.ThreadID,
		Direction: v.Direction,
		Message:   v.Message,
		Metadata:  json.RawMessage(v.Metadata),
		CreatedAt: v.CreatedAt,
	}
}
