package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/unknowncode44/appointments/internal/api/response"
	db "github.com/unknowncode44/appointments/internal/db/sqlc"
)

// Conversation FSM steps stored in conversation_state.state.
const (
	stepGreeting       = "greeting"
	stepChooseService  = "choose_service"
	stepChooseProvider = "choose_provider"
	stepChooseDate     = "choose_date"
	stepChooseSlot     = "choose_slot"
	stepConfirm        = "confirm"
	stepBooked         = "booked"
)

const (
	botListLimit = 50
	botDateDays  = 7 // how many upcoming days to offer
	anyProvider  = "any"
)

// botData is the working selection carried across turns in conversation_state.data.
// Offered holds the currently displayed numbered menu so the next inbound number
// maps back to a concrete id without re-querying.
type botData struct {
	ServiceID    string        `json:"service_id,omitempty"`
	ServiceName  string        `json:"service_name,omitempty"`
	ProviderID   string        `json:"provider_id,omitempty"` // "" or "any" = any provider
	ProviderName string        `json:"provider_name,omitempty"`
	Date         string        `json:"date,omitempty"` // YYYY-MM-DD
	SlotID       string        `json:"slot_id,omitempty"`
	Offered      []offeredItem `json:"offered,omitempty"`
}

type offeredItem struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type botInput struct {
	TenantID   uuid.UUID
	CustomerID uuid.UUID
	Step       string
	Data       botData
	Text       string
}

type botOutput struct {
	Reply string
	Step  string
	Data  botData
}

// botRepo is the read/booking surface the FSM needs. Kept small so unit tests can
// stub it without a database.
type botRepo interface {
	GetTenant(context.Context, uuid.UUID) (db.Tenant, error)
	ListServices(context.Context, db.ListServicesParams) ([]db.Service, error)
	ListProviders(context.Context, db.ListProvidersParams) ([]db.Provider, error)
	ListAvailableSlots(context.Context, db.ListAvailableSlotsParams) ([]db.AppointmentSlot, error)
	BookSlot(context.Context, bookSlotParams) (db.Appointment, error)
}

// bot is the conversational booking state machine. It is pure with respect to
// persistence: it receives the prior step/data and returns the reply plus the
// next step/data; the caller loads and saves conversation_state.
type bot struct {
	repo botRepo
	now  func() time.Time
}

func newBot(repo botRepo) *bot {
	return &bot{repo: repo, now: time.Now}
}

// Handle advances the conversation one step.
func (b *bot) Handle(ctx context.Context, in botInput) (botOutput, error) {
	text := strings.ToLower(strings.TrimSpace(in.Text))

	// Global reset: a greeting word, an empty/unknown step, or a completed booking
	// all start a fresh conversation.
	if in.Step == "" || in.Step == stepBooked || isResetWord(text) {
		return b.startGreeting(ctx, in.TenantID)
	}

	switch in.Step {
	case stepChooseService:
		return b.handleChooseService(ctx, in, text)
	case stepChooseProvider:
		return b.handleChooseProvider(ctx, in, text)
	case stepChooseDate:
		return b.handleChooseDate(ctx, in, text)
	case stepChooseSlot:
		return b.handleChooseSlot(ctx, in, text)
	case stepConfirm:
		return b.handleConfirm(ctx, in, text)
	default:
		return b.startGreeting(ctx, in.TenantID)
	}
}

// ── Steps ──────────────────────────────────────────────────────────────────

func (b *bot) startGreeting(ctx context.Context, tenantID uuid.UUID) (botOutput, error) {
	active := true
	services, err := b.repo.ListServices(ctx, db.ListServicesParams{TenantID: tenantID, Active: &active, Limit: botListLimit})
	if err != nil {
		return botOutput{}, err
	}
	if len(services) == 0 {
		return botOutput{Reply: "Por ahora no hay servicios disponibles para reservar. 🙏", Step: stepGreeting}, nil
	}
	offered := make([]offeredItem, len(services))
	for i, s := range services {
		offered[i] = offeredItem{ID: s.ID.String(), Label: s.Name}
	}
	reply := "¡Hola! 👋 Te ayudo a reservar un turno.\n\n" + renderMenu("¿Qué servicio querés?", offered)
	return botOutput{Reply: reply, Step: stepChooseService, Data: botData{Offered: offered}}, nil
}

func (b *bot) handleChooseService(ctx context.Context, in botInput, text string) (botOutput, error) {
	choice, ok := parseChoice(text, in.Data.Offered)
	if !ok {
		return reprompt(in, "No entendí. "+renderMenu("¿Qué servicio querés?", in.Data.Offered)), nil
	}
	data := in.Data
	data.ServiceID = choice.ID
	data.ServiceName = choice.Label
	return b.offerProviders(ctx, in.TenantID, data)
}

// offerProviders lists providers. With a single provider it auto-selects and
// jumps to dates; with several it offers them plus an "any" option.
func (b *bot) offerProviders(ctx context.Context, tenantID uuid.UUID, data botData) (botOutput, error) {
	active := true
	providers, err := b.repo.ListProviders(ctx, db.ListProvidersParams{TenantID: tenantID, Active: &active, Limit: botListLimit})
	if err != nil {
		return botOutput{}, err
	}
	if len(providers) == 1 {
		data.ProviderID = providers[0].ID.String()
		data.ProviderName = providers[0].Name
		return b.offerDates(ctx, tenantID, data)
	}
	offered := make([]offeredItem, 0, len(providers)+1)
	for _, p := range providers {
		offered = append(offered, offeredItem{ID: p.ID.String(), Label: p.Name})
	}
	offered = append(offered, offeredItem{ID: anyProvider, Label: "Cualquier profesional"})
	data.Offered = offered
	return botOutput{Reply: renderMenu("¿Con qué profesional?", offered), Step: stepChooseProvider, Data: data}, nil
}

func (b *bot) handleChooseProvider(ctx context.Context, in botInput, text string) (botOutput, error) {
	choice, ok := parseChoice(text, in.Data.Offered)
	if !ok {
		return reprompt(in, "No entendí. "+renderMenu("¿Con qué profesional?", in.Data.Offered)), nil
	}
	data := in.Data
	if choice.ID == anyProvider {
		data.ProviderID = ""
		data.ProviderName = ""
	} else {
		data.ProviderID = choice.ID
		data.ProviderName = choice.Label
	}
	return b.offerDates(ctx, in.TenantID, data)
}

func (b *bot) offerDates(ctx context.Context, tenantID uuid.UUID, data botData) (botOutput, error) {
	loc, err := b.tenantLoc(ctx, tenantID)
	if err != nil {
		return botOutput{}, err
	}
	today := b.now().In(loc)
	offered := make([]offeredItem, 0, botDateDays)
	for i := 0; i < botDateDays; i++ {
		d := today.AddDate(0, 0, i)
		offered = append(offered, offeredItem{ID: d.Format("2006-01-02"), Label: dayLabel(d)})
	}
	data.Offered = offered
	return botOutput{Reply: renderMenu("¿Qué día te viene bien?", offered), Step: stepChooseDate, Data: data}, nil
}

func (b *bot) handleChooseDate(ctx context.Context, in botInput, text string) (botOutput, error) {
	choice, ok := parseChoice(text, in.Data.Offered)
	if !ok {
		return reprompt(in, "No entendí. "+renderMenu("¿Qué día te viene bien?", in.Data.Offered)), nil
	}
	data := in.Data
	data.Date = choice.ID
	return b.offerSlots(ctx, in.TenantID, data)
}

func (b *bot) offerSlots(ctx context.Context, tenantID uuid.UUID, data botData) (botOutput, error) {
	loc, err := b.tenantLoc(ctx, tenantID)
	if err != nil {
		return botOutput{}, err
	}
	start, end, err := dayBounds(data.Date, loc)
	if err != nil {
		return botOutput{}, response.ErrInvalidInput
	}
	slots, err := b.repo.ListAvailableSlots(ctx, db.ListAvailableSlotsParams{
		TenantID:   tenantID,
		ProviderID: providerPtr(data.ProviderID),
		StartAt:    start,
		EndAt:      end,
		Limit:      botListLimit,
	})
	if err != nil {
		return botOutput{}, err
	}
	if len(slots) == 0 {
		// Re-offer dates so the user can pick another day.
		out, derr := b.offerDates(ctx, tenantID, data)
		if derr != nil {
			return botOutput{}, derr
		}
		out.Reply = "No hay turnos disponibles ese día 😕.\n\n" + out.Reply
		return out, nil
	}
	offered := make([]offeredItem, len(slots))
	for i, s := range slots {
		offered[i] = offeredItem{ID: s.ID.String(), Label: s.StartAt.In(loc).Format("15:04")}
	}
	data.Offered = offered
	return botOutput{Reply: renderMenu("Estos son los horarios disponibles:", offered), Step: stepChooseSlot, Data: data}, nil
}

func (b *bot) handleChooseSlot(ctx context.Context, in botInput, text string) (botOutput, error) {
	choice, ok := parseChoice(text, in.Data.Offered)
	if !ok {
		return reprompt(in, "No entendí. "+renderMenu("Estos son los horarios disponibles:", in.Data.Offered)), nil
	}
	data := in.Data
	data.SlotID = choice.ID
	summary := fmt.Sprintf("Confirmás tu turno de *%s*%s el %s a las %s?",
		data.ServiceName, providerSuffix(data.ProviderName), data.Date, choice.Label)
	return botOutput{Reply: summary + "\n\n1. Sí, confirmar\n2. No", Step: stepConfirm, Data: data}, nil
}

func (b *bot) handleConfirm(ctx context.Context, in botInput, text string) (botOutput, error) {
	switch {
	case isYes(text):
		return b.book(ctx, in)
	case isNo(text):
		return botOutput{Reply: "Sin problema, cancelé la reserva. Escribime *hola* cuando quieras empezar de nuevo. 👋", Step: stepBooked}, nil
	default:
		return reprompt(in, "Respondé *1* para confirmar o *2* para cancelar."), nil
	}
}

func (b *bot) book(ctx context.Context, in botInput) (botOutput, error) {
	serviceID, err := uuid.Parse(in.Data.ServiceID)
	if err != nil {
		return b.startGreeting(ctx, in.TenantID)
	}
	slotID, err := uuid.Parse(in.Data.SlotID)
	if err != nil {
		return b.startGreeting(ctx, in.TenantID)
	}

	_, err = b.repo.BookSlot(ctx, bookSlotParams{
		TenantID:   in.TenantID,
		CustomerID: in.CustomerID,
		ServiceID:  serviceID,
		SlotID:     slotID,
	})
	switch {
	case errors.Is(err, response.ErrConflict), errors.Is(err, response.ErrNotFound):
		// Slot was taken between offer and confirm — re-offer the same day's slots.
		out, oerr := b.offerSlots(ctx, in.TenantID, in.Data)
		if oerr != nil {
			return botOutput{}, oerr
		}
		out.Reply = "Uy, ese horario se ocupó 😕.\n\n" + out.Reply
		return out, nil
	case err != nil:
		return botOutput{}, err
	}

	reply := fmt.Sprintf("¡Listo! ✅ Tu turno de *%s*%s quedó confirmado para el %s a las %s.\n\n¡Te esperamos! 🙌",
		in.Data.ServiceName, providerSuffix(in.Data.ProviderName), in.Data.Date, slotLabel(in.Data))
	return botOutput{Reply: reply, Step: stepBooked}, nil
}

// ── Helpers ────────────────────────────────────────────────────────────────

func (b *bot) tenantLoc(ctx context.Context, tenantID uuid.UUID) (*time.Location, error) {
	tenant, err := b.repo.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	loc, lerr := time.LoadLocation(tenant.Timezone)
	if lerr != nil {
		return time.UTC, nil
	}
	return loc, nil
}

func renderMenu(prompt string, offered []offeredItem) string {
	lines := make([]string, 0, len(offered)+3)
	lines = append(lines, prompt, "")
	for i, o := range offered {
		lines = append(lines, fmt.Sprintf("%d. %s", i+1, o.Label))
	}
	lines = append(lines, "", "Respondé con el número.")
	return strings.Join(lines, "\n")
}

// parseChoice maps a numeric reply (1-based) onto an offered item.
func parseChoice(text string, offered []offeredItem) (offeredItem, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(text))
	if err != nil || n < 1 || n > len(offered) {
		return offeredItem{}, false
	}
	return offered[n-1], true
}

func reprompt(in botInput, reply string) botOutput {
	return botOutput{Reply: reply, Step: in.Step, Data: in.Data}
}

func isResetWord(text string) bool {
	switch text {
	case "hola", "menu", "menú", "inicio", "reiniciar", "empezar", "buenas":
		return true
	}
	return false
}

func isYes(text string) bool {
	switch text {
	case "si", "sí", "s", "1", "confirmar", "ok", "dale", "listo":
		return true
	}
	return false
}

func isNo(text string) bool {
	switch text {
	case "no", "n", "2", "cancelar":
		return true
	}
	return false
}

func providerPtr(id string) *uuid.UUID {
	if id == "" || id == anyProvider {
		return nil
	}
	p, err := uuid.Parse(id)
	if err != nil {
		return nil
	}
	return &p
}

func providerSuffix(name string) string {
	if name == "" {
		return ""
	}
	return " con " + name
}

func slotLabel(data botData) string {
	for _, o := range data.Offered {
		if o.ID == data.SlotID {
			return o.Label
		}
	}
	return ""
}

func dayLabel(d time.Time) string {
	days := [...]string{"dom", "lun", "mar", "mié", "jue", "vie", "sáb"}
	return fmt.Sprintf("%s %02d/%02d", days[d.Weekday()], d.Day(), int(d.Month()))
}

// dayBounds returns the [start, end) UTC instants for the given YYYY-MM-DD date
// interpreted in loc.
func dayBounds(date string, loc *time.Location) (time.Time, time.Time, error) {
	start, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start, start.AddDate(0, 0, 1), nil
}

// storeBotRepo adapts *db.Store to botRepo, running the shared booking core in a
// transaction.
type storeBotRepo struct{ store *db.Store }

func (r storeBotRepo) GetTenant(ctx context.Context, id uuid.UUID) (db.Tenant, error) {
	return r.store.GetTenant(ctx, id)
}
func (r storeBotRepo) ListServices(ctx context.Context, arg db.ListServicesParams) ([]db.Service, error) {
	return r.store.ListServices(ctx, arg)
}
func (r storeBotRepo) ListProviders(ctx context.Context, arg db.ListProvidersParams) ([]db.Provider, error) {
	return r.store.ListProviders(ctx, arg)
}
func (r storeBotRepo) ListAvailableSlots(ctx context.Context, arg db.ListAvailableSlotsParams) ([]db.AppointmentSlot, error) {
	return r.store.ListAvailableSlots(ctx, arg)
}
func (r storeBotRepo) BookSlot(ctx context.Context, p bookSlotParams) (db.Appointment, error) {
	var appt db.Appointment
	err := r.store.ExecTx(ctx, func(q *db.Queries) error {
		a, err := bookSlot(ctx, q, p)
		if err != nil {
			return err
		}
		appt = a
		return nil
	})
	return appt, err
}
