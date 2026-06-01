package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mousav1/ticket/internal/api/dto"
	"github.com/mousav1/ticket/internal/api/middleware"
	db "github.com/mousav1/ticket/internal/db/sqlc"
	"github.com/mousav1/ticket/internal/token"
	"github.com/mousav1/ticket/internal/util"
)

type UserHandler struct {
	Store      *db.Store
	tokenMaker token.Maker
	Config     util.Config
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
}

type UpdateUserRequest struct {
	FullName string `json:"full_name" binding:"required"`
}

type UpdatePasswordRequest struct {
	Password string `json:"password" binding:"required,min=6"`
}

type userResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

type loginUserResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

var validate = validator.New()

func NewUserHandler(Store *db.Store, tokenMaker token.Maker, Config util.Config) *UserHandler {
	return &UserHandler{
		Store,
		tokenMaker,
		Config,
	}
}

func (h *UserHandler) RegisterUser(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Validate query parameters
	if err := validate.Struct(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not hash password"})
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
	}

	user, err := h.Store.CreateUser(c.Context(), arg)

	if err != nil {
		// Assuming db.ErrorCode(err) is used to get the specific error code
		if db.ErrorCode(err) == db.UniqueViolation {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "username already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create user"})
	}

	rsp := newUserResponse(user)
	return c.Status(fiber.StatusCreated).JSON(rsp)
}

func (h *UserHandler) LoginUser(c *fiber.Ctx) error {
	var req loginUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := validate.Struct(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	user, err := h.Store.GetUserByUsernameFull(c.Context(), req.Username)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid username or password"})
	}

	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid username or password"})
	}

	accessToken, accessPayload, err := h.tokenMaker.CreateToken(user.Username, string(user.Role), user.TenantID, h.Config.AccessTokenDuration)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create token"})
	}

	refreshToken, refreshPayload, err := h.tokenMaker.CreateToken(user.Username, string(user.Role), user.TenantID, h.Config.RefreshTokenDuration)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create token"})
	}

	session, err := h.Store.CreateSession(c.Context(), db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    c.Get("User-Agent"),
		ClientIp:     c.IP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "could not create session"})
	}

	rsp := loginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  newUserResponse(user),
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"response": rsp})
}

// GetUserProfile handles fetching user profile
func (h *UserHandler) GetUserProfile(c *fiber.Ctx) error {
	payload := c.Locals(middleware.AuthorizationPayloadKey).(*token.Payload)

	user, err := h.Store.GetUserByUsernameFull(c.Context(), payload.Username)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"user": user})
}

func (h *UserHandler) UpdateUserProfile(c *fiber.Ctx) error {
	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := validate.Struct(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	payload := c.Locals(middleware.AuthorizationPayloadKey).(*token.Payload)

	user, err := h.Store.GetUserByUsername(c.Context(), payload.Username)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	arg := db.UpdateUserParams{
		Username: user.Username,
		FullName: req.FullName,
	}

	newUser, err := h.Store.UpdateUser(c.Context(), arg)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not update user"})
	}

	rsp := newUserResponse(newUser)
	return c.Status(fiber.StatusOK).JSON(rsp)
}

func (h *UserHandler) ChangePassword(c *fiber.Ctx) error {
	var req UpdatePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := validate.Struct(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not hash password"})
	}

	payload := c.Locals(middleware.AuthorizationPayloadKey).(*token.Payload)

	user, err := h.Store.GetUserByUsername(c.Context(), payload.Username)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	arg := db.UpdateUserPasswordParams{
		HashedPassword: hashedPassword,
		Username:       user.Username,
	}

	newUser, err := h.Store.UpdateUserPassword(c.Context(), arg)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not update password"})
	}

	rsp := newUserResponse(newUser)
	return c.Status(fiber.StatusOK).JSON(rsp)
}

// ListUsers fetches all users (admin only)
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	limit := 10
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	users, err := h.Store.ListUsers(c.Context(), db.ListUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch users"})
	}

	response := make([]dto.UserResponse, 0, len(users))
	for _, u := range users {
		response = append(response, dto.UserResponse{
			ID:        u.ID,
			Username:  u.Username,
			FullName:  u.FullName,
			Role:      string(u.Role),
			TenantID:  u.TenantID,
			CreatedAt: u.CreatedAt,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetUserByID fetches a specific user (admin only)
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	user, err := h.Store.GetUserByIDFull(c.Context(), int32(userID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	response := dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		FullName:  user.FullName,
		Role:      string(user.Role),
		TenantID:  user.TenantID,
		CreatedAt: user.CreatedAt,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// CreateUserAdmin creates a new user with role (admin only)
func (h *UserHandler) CreateUserAdmin(c *fiber.Ctx) error {
	var req dto.UserCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := validate.Struct(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not hash password"})
	}

	arg := db.CreateUserWithRoleParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Role:           db.UserRole(req.Role),
		TenantID:       req.TenantID,
	}

	user, err := h.Store.CreateUserWithRole(c.Context(), arg)
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "username already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create user"})
	}

	response := dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		FullName:  user.FullName,
		Role:      string(user.Role),
		TenantID:  user.TenantID,
		CreatedAt: user.CreatedAt,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// UpdateUser updates a user's role and tenant (admin only)
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	var req dto.UserUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	arg := db.UpdateUserRoleParams{
		ID:       int32(userID),
		Role:     db.UserRole(req.Role),
		TenantID: req.TenantID,
	}

	user, err := h.Store.UpdateUserRole(c.Context(), arg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not update user"})
	}

	response := dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		FullName:  user.FullName,
		Role:      string(user.Role),
		TenantID:  user.TenantID,
		CreatedAt: user.CreatedAt,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// DeleteUser deletes a user (admin only, cascades all related data)
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	err = h.Store.DeleteUser(c.Context(), int32(userID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not delete user"})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// LinkUserToTenant links a user to a tenant (admin only)
func (h *UserHandler) LinkUserToTenant(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	var req dto.UserTenantRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := validate.Struct(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	arg := db.UpdateUserTenantParams{
		ID:       int32(userID),
		TenantID: &req.TenantID,
	}

	user, err := h.Store.UpdateUserTenant(c.Context(), arg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not link user to tenant"})
	}

	response := dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		FullName:  user.FullName,
		Role:      string(user.Role),
		TenantID:  user.TenantID,
		CreatedAt: user.CreatedAt,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// LinkUserToProvider links a user to a provider (admin + tenantUser of same tenant)
func (h *UserHandler) LinkUserToProvider(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	var req dto.UserProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := validate.Struct(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	arg := db.AddUserProviderParams{
		UserID:     int32(userID),
		ProviderID: req.ProviderID,
	}

	err = h.Store.AddUserProvider(c.Context(), arg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not link user to provider"})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// RemoveUserFromProvider removes a user from a provider (admin + tenantUser of same tenant)
func (h *UserHandler) RemoveUserFromProvider(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	providerID := c.Query("provider_id")
	if providerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "provider_id is required"})
	}

	pID, err := uuid.Parse(providerID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid provider id"})
	}

	arg := db.RemoveUserProviderParams{
		UserID:     int32(userID),
		ProviderID: pID,
	}

	err = h.Store.RemoveUserProvider(c.Context(), arg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not remove user from provider"})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// GetUserProviders lists all providers for a user
func (h *UserHandler) GetUserProviders(c *fiber.Ctx) error {
	userID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	providers, err := h.Store.ListUserProviders(c.Context(), int32(userID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch providers"})
	}

	return c.Status(fiber.StatusOK).JSON(providers)
}
