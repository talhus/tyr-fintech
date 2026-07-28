package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iamtbay/tyr-fintech/internal/dto"
	jwtPkg "github.com/iamtbay/tyr-fintech/pkg/jwt"
	"github.com/iamtbay/tyr-fintech/pkg/response"
	"github.com/iamtbay/tyr-fintech/pkg/utils"
)

type UserService interface {
	Register(ctx context.Context, req *dto.RegisterUserRequest) error
	Login(ctx context.Context, req *dto.LoginUserRequest) (*dto.LoginResponse, error)
}

type UserHandler struct {
	userService UserService
}

// CREATE HANDLER
func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// REGISTER
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	if err := h.userService.Register(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// LOGIN
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	//JWT

	token, err := jwtPkg.GenerateToken(res.User.ID)
	utils.SetAuthCookie(c, token)

	response.Success(c, http.StatusOK, res)
}

// LOGOUT
func (h *UserHandler) Logout(c *gin.Context) {
	utils.ClearAuthCookie(c)
	c.JSON(http.StatusOK, gin.H{"message": "Logout success"})
}
