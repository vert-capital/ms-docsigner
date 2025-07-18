package handlers

import (
	"app/entity"
	"app/infrastructure/repository"
	usecase_user "app/usecase/user"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserPasswordData struct {
	Email           string `json:"email"`
	OldPassword     string `json:"oldPassword"`
	NewPassword     string `json:"newPassword"`
	ConfirmPassword string `json:"confirmPassword"`
}

type UserHandlers struct {
	UsecaseUser usecase_user.IUsecaseUser
}

func NewUserHandler(usecaseUser usecase_user.IUsecaseUser) *UserHandlers {
	return &UserHandlers{UsecaseUser: usecaseUser}
}

// @Summary Login
// @Description Login
// @Tags User
// @Accept  json
// @Produce  json
// @Param email body string true "Email"
// @Param password body string true "Password"
// @Success 200 {object} entity.EntityUser "success"
// @Router /api/login [post]
func (h UserHandlers) LoginHandler(c *gin.Context) {

	var loginData LoginData

	if err := c.ShouldBindJSON(&loginData); err != nil {
		handleError(c, err)
		return
	}

	user, err := h.UsecaseUser.LoginUser(loginData.Email, loginData.Password)

	if exception := handleError(c, err); exception {
		return
	}

	token, refreshToken, err := usecase_user.JWTTokenGenerator(*user)

	if exception := handleError(c, err); exception {
		return
	}

	jsonResponse(c, http.StatusOK, gin.H{"token": token, "refreshToken": refreshToken})
}

// @Summary Get me
// @Description Get me
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} entity.EntityUser "success"
// @Router /api/user/me [get]
func (h UserHandlers) GetMeHandler(c *gin.Context) {
	user, err := h.UsecaseUser.GetUserByToken(c.GetHeader("Authorization"))

	if exception := handleError(c, err); exception {
		return
	}

	jsonResponse(c, http.StatusOK, user)
}

// @Summary Create user
// @Description Create user
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param entity.EntityUser body entity.EntityUser true "User"
// @Success 200 {object} entity.EntityUser "success"
// @Router /api/user/create [post]
func (h UserHandlers) CreateUserHandler(c *gin.Context) {

	var entityUser entity.EntityUser

	if err := c.ShouldBindJSON(&entityUser); err != nil {
		handleError(c, err)
		return
	}

	err := h.UsecaseUser.Create(&entityUser)

	if exception := handleError(c, err); exception {
		return
	}

	jsonResponse(c, http.StatusOK, gin.H{"message": "User created successfully"})
}

// @Summary Update user
// @Description Update user
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "User ID"
// @Param entity.EntityUser body entity.EntityUser true "User"
// @Success 200 {object} entity.EntityUser "success"
// @Router /api/user/{id} [put]
func (h UserHandlers) UpdateUserHandler(c *gin.Context) {

	var entityUser entity.EntityUser

	id := strconv.Itoa(c.GetInt("id"))

	dataInt, _ := strconv.Atoi(id)

	entityUser.ID = dataInt

	if err := c.ShouldBindJSON(&entityUser); err != nil {
		handleError(c, err)
		return
	}

	err := h.UsecaseUser.Update(&entityUser)

	if exception := handleError(c, err); exception {
		return
	}

	jsonResponse(c, http.StatusOK, gin.H{"message": "User updated successfully"})
}

// @Summary Delete user
// @Description Delete user
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "User ID"
// @Success 200 {object} entity.EntityUser "success"
// @Router /api/user/{id} [delete]
func (h UserHandlers) DeleteUserHandler(c *gin.Context) {

	var entityUser entity.EntityUser

	if err := c.ShouldBindJSON(&entityUser); err != nil {
		handleError(c, err)
		return
	}

	err := h.UsecaseUser.Delete(&entityUser)

	if exception := handleError(c, err); exception {
		return
	}

	jsonResponse(c, http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// @Summary Update password
// @Description Update password
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "User ID"
// @Param entity.EntityUser body entity.EntityUser true "User"
// @Success 200 {object} entity.EntityUser "success"
// @Router /api/user/password/{id} [put]
func (h UserHandlers) UpdatePasswordHandler(c *gin.Context) {

	var updatePasswordData UpdateUserPasswordData

	if err := c.ShouldBindJSON(&updatePasswordData); err != nil {
		handleError(c, err)
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))

	err := h.UsecaseUser.UpdatePassword(id, updatePasswordData.OldPassword, updatePasswordData.NewPassword, updatePasswordData.ConfirmPassword)

	if exception := handleError(c, err); exception {
		return
	}

	jsonResponse(c, http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// @Summary Get users
// @Description Get users
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param search query string false "Search"
// @Param active query string false "Active"
// @Success 200 {object} entity.EntityUser "success"
// @Router /api/user/list [get]
func (h UserHandlers) GetUsersHandler(c *gin.Context) {

	var filters entity.EntityUserFilters

	filters.Search = c.Query("search")
	filters.Active = c.Query("active")

	users, err := h.UsecaseUser.GetUsers(filters)

	if exception := handleError(c, err); exception {
		return
	}

	jsonResponse(c, http.StatusOK, users)
}

// @Summary Get user
// @Description Get user
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "User ID"
// @Success 200 {object} entity.EntityUser "success"
// @Router /api/user/{id} [get]
func (h UserHandlers) GetUserHandler(c *gin.Context) {

	id, _ := strconv.Atoi(c.Param("id"))

	user, err := h.UsecaseUser.GetUser(id)

	if exception := handleError(c, err); exception {
		return
	}

	jsonResponse(c, http.StatusOK, user)
}

func MountUsersHandlers(gin *gin.Engine, conn *gorm.DB) {

	userHandlers := NewUserHandler(
		usecase_user.NewService(
			repository.NewUserPostgres(conn),
		),
	)

	gin.GET("/", HomeHandler)
	gin.POST("/api/login", userHandlers.LoginHandler)

	gin.POST("/login", userHandlers.LoginHandler)

	// user
	group := gin.Group("/api/user")
	SetAuthMiddleware(conn, group)

	group.GET("/me", userHandlers.GetMeHandler)
	group.POST("/create", userHandlers.CreateUserHandler)
	group.PUT("/:id", userHandlers.UpdateUserHandler)
	group.DELETE("/:id", userHandlers.DeleteUserHandler)
	group.PUT("/password/:id", userHandlers.UpdatePasswordHandler)
	group.GET("/list", userHandlers.GetUsersHandler)
	group.GET("/:id", userHandlers.GetUserHandler)
}
