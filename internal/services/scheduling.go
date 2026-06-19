package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/unknowncode44/appointments/internal/api/dto"
	"github.com/unknowncode44/appointments/internal/api/response"
	db "github.com/unknowncode44/appointments/internal/db/sqlc"
	"github.com/unknowncode44/appointments/internal/platform/pagination"
	"github.com/unknowncode44/appointments/internal/repositories"
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
	item, err := s.repo.CreateAvailability(ctx, db.CreateProviderAvailabilityParams{
		ProviderID: providerID,
		Weekday:    req.Weekday,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
	})
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
	item, err := s.repo.CreateException(ctx, db.CreateProviderExceptionParams{
		ProviderID: providerID,
		StartAt:    req.StartAt,
		EndAt:      req.EndAt,
		Reason:     repositories.Text(req.Reason),
	})
	return mapException(item), err
}

func (s *schedulingService) ListExceptions(ctx context.Context, providerID uuid.UUID) ([]dto.ExceptionResponse, error) {
	items, err := s.repo.ListExceptions(ctx, providerID)
	return mapSlice(items, mapException), err
}

// GenerateSlots creates appointment slots for a provider over the requested
// number of weeks.
//
// Timezone: slots are generated in the tenant's local timezone (loaded from
// req.Timezone). For Argentine tenants this should be
// "America/Argentina/Buenos_Aires". If the timezone is empty or invalid the
// function falls back to UTC and continues (non-fatal).
//
// Error handling: unique-constraint violations are skipped silently (the slot
// already exists). Any other DB error is propagated immediately — the caller
// learns about the failure instead of receiving a wrong count.
func (s *schedulingService) GenerateSlots(ctx context.Context, req dto.SlotGeneratorRequest) (dto.SlotGeneratorResponse, error) {
	slotMinutes := req.SlotMinutes
	if slotMinutes <= 0 {
		slotMinutes = 30
	}

	// Resolve tenant timezone — default UTC if not set or unrecognised.
	loc := time.UTC
	if req.Timezone != "" {
		if l, err := time.LoadLocation(req.Timezone); err == nil {
			loc = l
		}
	}

	availability, err := s.repo.ListAvailability(ctx, req.ProviderID)
	if err != nil {
		return dto.SlotGeneratorResponse{}, err
	}

	// Use the tenant's local "today" as the start boundary.
	now := time.Now().In(loc)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	end := start.AddDate(0, 0, req.NumberOfWeeks*7)

	exceptions, err := s.repo.ListExceptionsBetween(ctx, db.ListProviderExceptionsBetweenParams{
		ProviderID: req.ProviderID,
		StartAt:    start,
		EndAt:      end,
	})
	if err != nil {
		return dto.SlotGeneratorResponse{}, err
	}

	generated := 0
	for day := start; day.Before(end); day = day.AddDate(0, 0, 1) {
		for _, av := range availability {
			if int(av.Weekday) != int(day.Weekday()) {
				continue
			}
			cursor := combineInLoc(day, av.StartTime.Format("15:04:05"), loc)
			periodEnd := combineInLoc(day, av.EndTime.Format("15:04:05"), loc)
			for cursor.Add(time.Duration(slotMinutes) * time.Minute).Compare(periodEnd) <= 0 {
				slotEnd := cursor.Add(time.Duration(slotMinutes) * time.Minute)
				if !overlapsAny(cursor, slotEnd, exceptions) {
					_, err := s.repo.CreateSlot(ctx, db.CreateAppointmentSlotParams{
						TenantID:   req.TenantID,
						ProviderID: req.ProviderID,
						StartAt:    cursor,
						EndAt:      slotEnd,
					})
					if err != nil {
						// Unique violation = slot already exists; skip quietly.
						if db.ErrorCode(err) == db.UniqueViolation {
							cursor = slotEnd
							continue
						}
						// Any other DB error is fatal.
						return dto.SlotGeneratorResponse{Generated: generated}, err
					}
					generated++
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
	items, err := s.repo.ListAvailableSlots(ctx, db.ListAvailableSlotsParams{
		TenantID:   tenantID,
		ProviderID: providerID,
		ServiceID:  serviceID,
		StartAt:    start,
		EndAt:      end,
		Limit:      p.PageSize,
		Offset:     p.Offset,
	})
	return mapSlice(items, mapSlot), err
}

// sameDayBounds parses a "2006-01-02" date string and returns the UTC start
// and end of that calendar day. Moved here from mappers.go (it's scheduling
// logic, not a data mapper).
func sameDayBounds(date string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start, start.Add(24 * time.Hour), nil
}

// combineInLoc builds a time.Time by combining the calendar date of `day` with
// the clock time parsed from `clockStr` ("HH:MM" or "HH:MM:SS") in location loc.
func combineInLoc(day time.Time, clockStr string, loc *time.Location) time.Time {
	t, err := time.Parse("15:04", clockStr)
	if err != nil {
		// Try HH:MM:SS fallback (DB may store seconds).
		t, err = time.Parse("15:04:05", clockStr)
		if err != nil {
			return day
		}
	}
	return time.Date(day.Year(), day.Month(), day.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc)
}

func overlapsAny(start, end time.Time, exceptions []db.ProviderException) bool {
	for _, ex := range exceptions {
		if start.Before(ex.EndAt) && end.After(ex.StartAt) {
			return true
		}
	}
	return false
}

// mapSlice is a generic helper shared across service packages.
func mapSlice[I any, O any](items []I, mapper func(I) O) []O {
	out := make([]O, 0, len(items))
	for _, item := range items {
		out = append(out, mapper(item))
	}
	return out
}

// pgxErrNoRows shadows the pgx import for internal use.
var _ = pgx.ErrNoRows
