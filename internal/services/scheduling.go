package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mousav1/ticket/internal/api/dto"
	"github.com/mousav1/ticket/internal/api/response"
	db "github.com/mousav1/ticket/internal/db/sqlc"
	"github.com/mousav1/ticket/internal/platform/pagination"
	"github.com/mousav1/ticket/internal/repositories"
)

type SchedulingService interface {
	CreateAvailability(context.Context, uuid.UUID, dto.AvailabilityRequest) (dto.AvailabilityResponse, error)
	ListAvailability(context.Context, uuid.UUID) ([]dto.AvailabilityResponse, error)
	CreateException(context.Context, uuid.UUID, dto.ExceptionRequest) (dto.ExceptionResponse, error)
	ListExceptions(context.Context, uuid.UUID) ([]dto.ExceptionResponse, error)
	GenerateSlots(context.Context, dto.SlotGeneratorRequest) (dto.SlotGeneratorResponse, error)
	ListAvailableSlots(context.Context, uuid.UUID, *uuid.UUID, *uuid.UUID, string, pagination.Page) ([]dto.SlotResponse, error)
}

type schedulingService struct {
	repo repositories.SchedulingRepository
}

func NewSchedulingService(repo repositories.SchedulingRepository) SchedulingService {
	return &schedulingService{repo: repo}
}

func (s *schedulingService) CreateAvailability(ctx context.Context, providerID uuid.UUID, req dto.AvailabilityRequest) (dto.AvailabilityResponse, error) {
	if _, err := time.Parse("15:04", req.StartTime); err != nil {
		return dto.AvailabilityResponse{}, response.ErrInvalidInput
	}
	if _, err := time.Parse("15:04", req.EndTime); err != nil {
		return dto.AvailabilityResponse{}, response.ErrInvalidInput
	}
	item, err := s.repo.CreateAvailability(ctx, db.CreateProviderAvailabilityParams{ProviderID: providerID, Weekday: req.Weekday, StartTime: req.StartTime, EndTime: req.EndTime})
	return mapAvailability(item), err
}

func (s *schedulingService) ListAvailability(ctx context.Context, providerID uuid.UUID) ([]dto.AvailabilityResponse, error) {
	items, err := s.repo.ListAvailability(ctx, providerID)
	return mapSlice(items, mapAvailability), err
}

func (s *schedulingService) CreateException(ctx context.Context, providerID uuid.UUID, req dto.ExceptionRequest) (dto.ExceptionResponse, error) {
	if !req.EndAt.After(req.StartAt) {
		return dto.ExceptionResponse{}, response.ErrInvalidInput
	}
	item, err := s.repo.CreateException(ctx, db.CreateProviderExceptionParams{ProviderID: providerID, StartAt: req.StartAt, EndAt: req.EndAt, Reason: repositories.Text(req.Reason)})
	return mapException(item), err
}

func (s *schedulingService) ListExceptions(ctx context.Context, providerID uuid.UUID) ([]dto.ExceptionResponse, error) {
	items, err := s.repo.ListExceptions(ctx, providerID)
	return mapSlice(items, mapException), err
}

func (s *schedulingService) GenerateSlots(ctx context.Context, req dto.SlotGeneratorRequest) (dto.SlotGeneratorResponse, error) {
	slotMinutes := req.SlotMinutes
	if slotMinutes <= 0 {
		slotMinutes = 30
	}
	availability, err := s.repo.ListAvailability(ctx, req.ProviderID)
	if err != nil {
		return dto.SlotGeneratorResponse{}, err
	}

	start := time.Now().Truncate(24 * time.Hour)
	end := start.AddDate(0, 0, req.NumberOfWeeks*7)
	exceptions, err := s.repo.ListExceptionsBetween(ctx, db.ListProviderExceptionsBetweenParams{ProviderID: req.ProviderID, StartAt: start, EndAt: end})
	if err != nil {
		return dto.SlotGeneratorResponse{}, err
	}

	generated := 0
	for day := start; day.Before(end); day = day.AddDate(0, 0, 1) {
		for _, av := range availability {
			if int(av.Weekday) != int(day.Weekday()) {
				continue
			}
			cursor := combine(day, av.StartTime)
			periodEnd := combine(day, av.EndTime)
			for cursor.Add(time.Duration(slotMinutes)*time.Minute).Compare(periodEnd) <= 0 {
				slotEnd := cursor.Add(time.Duration(slotMinutes) * time.Minute)
				if !overlapsAny(cursor, slotEnd, exceptions) {
					_, err := s.repo.CreateSlot(ctx, db.CreateAppointmentSlotParams{TenantID: req.TenantID, ProviderID: req.ProviderID, StartAt: cursor, EndAt: slotEnd})
					if err == nil {
						generated++
					} else if !errors.Is(err, pgx.ErrNoRows) {
						return dto.SlotGeneratorResponse{}, err
					}
				}
				cursor = slotEnd
			}
		}
	}
	return dto.SlotGeneratorResponse{Generated: generated}, nil
}

func (s *schedulingService) ListAvailableSlots(ctx context.Context, tenantID uuid.UUID, providerID *uuid.UUID, serviceID *uuid.UUID, date string, p pagination.Page) ([]dto.SlotResponse, error) {
	start, end, err := sameDayBounds(date)
	if err != nil {
		return nil, response.ErrInvalidInput
	}
	items, err := s.repo.ListAvailableSlots(ctx, db.ListAvailableSlotsParams{TenantID: tenantID, ProviderID: providerID, ServiceID: serviceID, StartAt: start, EndAt: end, Limit: p.PageSize, Offset: p.Offset})
	return mapSlice(items, mapSlot), err
}

func combine(day time.Time, clock time.Time) time.Time {
	return time.Date(day.Year(), day.Month(), day.Day(), clock.Hour(), clock.Minute(), clock.Second(), 0, day.Location())
}

func overlapsAny(start, end time.Time, exceptions []db.ProviderException) bool {
	for _, ex := range exceptions {
		if start.Before(ex.EndAt) && end.After(ex.StartAt) {
			return true
		}
	}
	return false
}

func mapSlice[I any, O any](items []I, mapper func(I) O) []O {
	out := make([]O, 0, len(items))
	for _, item := range items {
		out = append(out, mapper(item))
	}
	return out
}
