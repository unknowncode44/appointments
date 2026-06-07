package handlers

import (
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mousav1/ticket/internal/api/dto"
	"github.com/mousav1/ticket/internal/api/middleware"
	"github.com/mousav1/ticket/internal/api/response"
	db "github.com/mousav1/ticket/internal/db/sqlc"
	"github.com/mousav1/ticket/internal/platform/pagination"
	"github.com/mousav1/ticket/internal/token"
	"github.com/mousav1/ticket/internal/util"
)

type UserHandler struct {
	Store      *db.Store
	tokenMaker token.Maker
	Config     util.Config
}

// CreateUserRequest uses validate tags (not binding) so go-playground/validator picks them up.
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,alphanum"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name" validate:"required"`
}

type UpdateUserRequest struct {
	FullName string `json:"full_name" validate:"required"`
}

type UpdatePasswordRequest struct {
	Password string `json:"password" validate:"required,min=6"`
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
	Username string `json:"username" validate:"required,alphanum"`
	Password string `json:"password" validate:"required,min=6"`
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
	return &UserHandler{Store, tokenMaker, Config}
}

func (h *UserHandler) RegisterUser(c *fiber.Ctx) error {
	var req CreateUserRequest
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
	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
	}
	user, err := h.Store.CreateUser(c.Context(), arg)
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "username already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create user"})
	}
	return c.Status(fiber.StatusCreated).JSON(newUserResponse(user))
}

func (h *UserHandler) LoginUser(c *fiber.Ctx) error {
	var req loginUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := validate.Struct(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	user, err := h.Store.GetUserByUsername(c.Context(), req.Username)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}
	if err = util.CheckPassword(req.Password, user.HashedPassword); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}
	accessToken, accessPayload, err := h.tokenMaker.CreateToken(user.Username, string(user.Role), user.TenantID, h.Config.AccessTokenDuration)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create token"})
	}
	refreshToken, refreshPayload, err := h.tokenMaker.CreateToken(user.Username, string(user.Role), user.TenantID, h.Config.RefreshTokenDuration)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create refresh token"})
	}
	session, err := h.Store.CreateSession(c.Context(), db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    string(c.Request().Header.UserAgent()),
		ClientIp:     c.IP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create session"})
	}
	return c.Status(fiber.StatusOK).JSON(loginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  newUserResponse(user),
	})
}

func (h *UserHandler) GetUserProfile(c *fiber.Ctx) error {
	payload, ok := c.Locals(middleware.AuthorizationPayloadKey).(*token.Payload)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid authorization payload"})
	}
	user, err := h.Store.GetUserByUsername(c.Context(), payload.Username)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(newUserResponse(user))
}

func (h *UserHandler) UpdateUserProfile(c *fiber.Ctx) error {
	payload, ok := c.Locals(middleware.AuthorizationPayloadKey).(*token.Payload)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid authorization payload"})
	}
	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := validate.Struct(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	user, err := h.Store.UpdateUserProfile(c.Context(), db.UpdateUserProfileParams{
		Username: payload.Username,
		FullName: req.FullName,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not update profile"})
	}
	return c.Status(fiber.StatusOK).JSON(newUserResponse(user))
}

func (h *UserHandler) ChangePassword(c *fiber.Ctx) error {
	payload, ok := c.Locals(middleware.AuthorizationPayloadKey).(*token.Payload)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid authorization payload"})
	}
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
	user, err := h.Store.GetUserByUsername(c.Context(), payload.Username)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	newUser, err := h.Store.UpdateUserPassword(c.Context(), db.UpdateUserPasswordParams{
		HashedPassword: hashedPassword,
		Username:       user.Username,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not update password"})
	}
	return c.Status(fiber.StatusOK).JSON(newUserResponse(newUser))
}

// ListUsers uses pagination.FromCtx for consistency with all other list endpoints.
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	p := pagination.FromCtx(c)
	users, err := h.Store.ListUsers(c.Context(), db.ListUsersParams{
		Limit:  p.PageSize,
		Offset: p.Offset,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch users"})
	}
	resp := make([]dto.UserResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, dto.UserResponse{
			ID:        u.ID,
			Username:  u.Username,
			FullName:  u.FullName,
			Role:      string(u.Role),
			TenantID:  u.TenantID,
			CreatedAt: u.CreatedAt,
		})
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

// GetUserByID uses uuid.UUID — consistent with all other domain entities.
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}
	user, err := h.Store.GetUserByIDFull(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	return c.Status(fiber.StatusOK).JSON(dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		FullName:  user.FullName,
		Role:      string(user.Role),
		TenantID:  user.TenantID,
		CreatedAt: user.CreatedAt,
	})
}

// CreateUserAdmin creates a user with explicit role (admin only).
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
	user, err := h.Store.CreateUserWithRole(c.Context(), db.CreateUserWithRoleParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Role:           db.UserRole(req.Role),
		TenantID:       req.TenantID,
	})
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "username already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create user"})
	}
	return c.Status(fiber.StatusCreated).JSON(dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		FullName:  user.FullName,
		Role:      string(user.Role),
		TenantID:  user.TenantID,
		CreatedAt: user.CreatedAt,
	})
}

// UpdateUser uses uuid.UUID for the path param.
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}
	var req dto.UserUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	user, err := h.Store.UpdateUserRole(c.Context(), db.UpdateUserRoleParams{
		ID:       userID,
		Role:     db.UserRole(req.Role),
		TenantID: req.TenantID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not update user"})
	}
	return c.Status(fiber.StatusOK).JSON(dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		FullName:  user.FullName,
		Role:      string(user.Role),
		TenantID:  user.TenantID,
		CreatedAt: user.CreatedAt,
	})
}

// DeleteUser uses uuid.UUID for the path param.
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}
	if err = h.Store.DeleteUser(c.Context(), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not delete user"})
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}

// LinkUserToTenant uses uuid.UUID for the path param.
func (h *UserHandler) LinkUserToTenant(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
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
	user, err := h.Store.UpdateUserTenant(c.Context(), db.UpdateUserTenantParams{
		ID:       userID,
		TenantID: &req.TenantID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not link user to tenant"})
	}
	return c.Status(fiber.StatusOK).JSON(dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		FullName:  user.FullName,
		Role:      string(user.Role),
		TenantID:  user.TenantID,
		CreatedAt: user.CreatedAt,
	})
}

// LinkUserToProvider uses uuid.UUID for the path param.
func (h *UserHandler) LinkUserToProvider(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
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
	if err = h.Store.AddUserProvider(c.Context(), db.AddUserProviderParams{
		UserID:     userID,
		ProviderID: req.ProviderID,
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not link user to provider"})
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}

// RemoveUserFromProvider uses uuid.UUID for the path param.
func (h *UserHandler) RemoveUserFromProvider(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}
	providerID, err := uuid.Parse(c.Query("provider_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid provider_id"})
	}
	if err = h.Store.RemoveUserProvider(c.Context(), db.RemoveUserProviderParams{
		UserID:     userID,
		ProviderID: providerID,
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not remove user from provider"})
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}

// GetUserProviders uses uuid.UUID for the path param.
func (h *UserHandler) GetUserProviders(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}
	providers, err := h.Store.ListUserProviders(c.Context(), userID)
	if err != nil {
		return response.Error(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(providers)
}
