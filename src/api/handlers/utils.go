package handlers

import (
	"app/api/middleware"
	"app/infrastructure/repository"
	usecase_user "app/usecase/user"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PaginationResponse struct {
	TotalPages     int `json:"total_pages"`
	Page           int `json:"page"`
	PageSize       int `json:"page_size"`
	TotalRegisters int `json:"total_registers"`
	Registers      any `json:"registers"`
}

func handleError(c *gin.Context, err error) bool {
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return true
	}
	return false
}

func jsonResponse(c *gin.Context, httpStatus int, data any) {
	c.JSON(httpStatus, data)
}

func RoutersHandler(c *gin.Context, r *gin.Engine) {
	type Router struct {
		Method string `json:"method"`
		Path   string `json:"path"`
	}

	var routers []Router = make([]Router, 0)

	for _, route := range r.Routes() {
		routers = append(routers, Router{
			Method: route.Method,
			Path:   route.Path,
		})
	}

	if gin.Mode() == gin.DebugMode {
		c.JSON(200, routers)
	}
}

func SetAuthMiddleware(conn *gorm.DB, group *gin.RouterGroup) {
	usecaseUser := usecase_user.NewService(
		repository.NewUserPostgres(conn),
	)

	group.Use(middleware.AuthenticatedMiddleware(usecaseUser))
}

func SetAdminMiddleware(conn *gorm.DB, group *gin.RouterGroup) {
	usecaseUser := usecase_user.NewService(
		repository.NewUserPostgres(conn),
	)

	group.Use(middleware.AdminMiddleware(usecaseUser))
}

func getPaginationParams(c *gin.Context) (int, int) {
	page := 0
	pageSize := 10

	if c.Query("page") != "" {
		page, _ = strconv.Atoi(c.Query("page"))
	}
	if c.Query("page_size") != "" {
		pageSize, _ = strconv.Atoi(c.Query("page_size"))
		if pageSize < 1 {
			pageSize = 10
		}

		if pageSize > 100 {
			pageSize = 100
		}
	}
	return page, pageSize
}

func getOrderAndSortByParams(c *gin.Context, defaultOrder string, defaultSort string) (string, string) {
	orderBy := c.Query("order_by")
	sortOrder := c.Query("sort_order")

	if orderBy == "" {
		orderBy = defaultOrder
	}

	if sortOrder == "" {
		sortOrder = defaultSort
	}

	return orderBy, sortOrder
}

func getOrderByParams(c *gin.Context, defaultValue string) (string, string) {
	return getOrderAndSortByParams(c, defaultValue, "asc")
}

func getTotalPaginas(totalRegistros int64, tamanhoPagina int) int {
	return int(math.Ceil(float64(totalRegistros) / float64(tamanhoPagina)))
}
