package v1

import (
	"github.com/foldn/bi-go/internal/models"
	"github.com/foldn/bi-go/internal/models/apierror"
	"github.com/foldn/bi-go/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type DataSourceHandler struct {
	service service.DataSourceService
}

func NewDataSourceHandler(s service.DataSourceService) *DataSourceHandler {
	return &DataSourceHandler{service: s}
}

// ErrorResponse represents a generic JSON error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// CreateDataSource godoc
// @Summary Create a new data source
// @Description Add a new data source configuration to the system
// @Tags datasources
// @Accept  json
// @Produce  json
// @Param   datasource  body   models.CreateDataSourceInput  true  "Data Source Configuration"
// @Success 201 {object} models.DataSource
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /datasources [post]
func (h *DataSourceHandler) CreateDataSource(c *gin.Context) {
	var input models.CreateDataSourceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(apierror.NewBadRequest("invalid input"))
		return
	}

	ds, err := h.service.CreateDataSource(input)
	if err != nil {
		// Example of more specific error handling
		if err.Error() == "datasource with this name already exists" {
			_ = c.Error(apierror.ErrConflict)
			return
		}
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, ds)
}

// GetDataSources godoc
// @Summary Get all data sources
// @Description Retrieve a paginated list of data sources
// @Tags datasources
// @Produce  json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Number of items per page" default(10)
// @Success 200 {object} map[string]interface{} "data, total, page, pageSize"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /datasources [get]
func (h *DataSourceHandler) GetDataSources(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 { // Max page size limit
		pageSize = 100
	}

	dataSources, total, err := h.service.GetDataSources(page, pageSize)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     dataSources,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetDataSourceByID godoc
// @Summary Get a data source by ID
// @Description Retrieve a specific data source configuration by its ID
// @Tags datasources
// @Produce  json
// @Param   id   path   int  true  "Data Source ID"
// @Success 200 {object} models.DataSource
// @Failure 404 {object} ErrorResponse "Data source not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /datasources/{id} [get]
func (h *DataSourceHandler) GetDataSourceByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		_ = c.Error(apierror.NewBadRequest("invalid id"))
		return
	}

	ds, err := h.service.GetDataSourceByID(uint(id))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ds)
}

// UpdateDataSource godoc
// @Summary Update an existing data source
// @Description Update an existing data source configuration by its ID
// @Tags datasources
// @Accept  json
// @Produce  json
// @Param   id   path   int  true  "Data Source ID"
// @Param   datasource  body   models.UpdateDataSourceInput  true  "Data Source Configuration Update"
// @Success 200 {object} models.DataSource
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 404 {object} ErrorResponse "Data source not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /datasources/{id} [put]
func (h *DataSourceHandler) UpdateDataSource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		_ = c.Error(apierror.NewBadRequest("invalid id"))
		return
	}

	var input models.UpdateDataSourceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(apierror.NewBadRequest("invalid input"))
		return
	}

	ds, err := h.service.UpdateDataSource(uint(id), input)
	if err != nil {
		if err.Error() == "datasource with this name already exists" {
			_ = c.Error(apierror.ErrConflict)
			return
		}

		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, ds)
}

// DeleteDataSource godoc
// @Summary Delete a data source
// @Description Delete a data source configuration by its ID
// @Tags datasources
// @Produce  json
// @Param   id   path   int  true  "Data Source ID"
// @Success 204 "Successfully deleted"
// @Failure 404 {object} ErrorResponse "Data source not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /datasources/{id} [delete]
func (h *DataSourceHandler) DeleteDataSource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		_ = c.Error(apierror.NewBadRequest("invalid id"))
		return
	}

	err = h.service.DeleteDataSource(uint(id))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetDataSourceSchema godoc
// @Summary Get schema of a data source
// @Description Retrieve the top-level schema (e.g., list of tables) of a data source
// @Tags datasources
// @Produce  json
// @Param   id   path   int  true  "Data Source ID"
// @Success 200 {object} interface{} "Schema Information"
// @Failure 400 {object} ErrorResponse "Invalid ID format"
// @Failure 404 {object} ErrorResponse "Data source not found"
// @Failure 500 {object} ErrorResponse "Error fetching schema"
// @Router /datasources/{id}/schema [get]
func (h *DataSourceHandler) GetDataSourceSchema(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		_ = c.Error(apierror.NewBadRequest("invalid id"))
		return
	}

	// Ensure the datasource exists first (optional, service might do this)
	_, err = h.service.GetDataSourceByID(uint(id))
	if err != nil {
		_ = c.Error(err)
		return
	}

	schema, err := h.service.GetDataSourceSchema(uint(id))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, schema)
}

// GetDataSourceEntitySchema godoc
// @Summary Get schema of a specific entity within a data source
// @Description Retrieve the schema (e.g., columns and types) of a specific table or entity
// @Tags datasources
// @Produce  json
// @Param   id   path   int  true  "Data Source ID"
// @Param   entity_name   path   string  true  "Entity Name (e.g., table name)"
// @Success 200 {object} interface{} "Entity Schema Information"
// @Failure 400 {object} ErrorResponse "Invalid ID or entity name"
// @Failure 404 {object} ErrorResponse "Data source or entity not found"
// @Failure 500 {object} ErrorResponse "Error fetching entity schema"
// @Router /datasources/{id}/schema/{entity_name} [get]
func (h *DataSourceHandler) GetDataSourceEntitySchema(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		_ = c.Error(apierror.NewBadRequest("invalid id"))
		return
	}

	entityName := c.Param("entity_name")
	if entityName == "" {
		_ = c.Error(apierror.NewBadRequest("entity name cannot be empty"))
		return
	}

	// Ensure the datasource exists first (optional, service might do this)
	_, err = h.service.GetDataSourceByID(uint(id))
	if err != nil {
		_ = c.Error(err)
		return
	}

	schema, err := h.service.GetDataSourceEntitySchema(uint(id), entityName)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, schema)
}
