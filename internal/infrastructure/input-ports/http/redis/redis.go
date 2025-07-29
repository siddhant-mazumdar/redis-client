package redis

import (
	"net/http"
	"strconv"
	"sync"

	"go-redis/internal/infrastructure/interface-adapters/sqlite"
	redis_usecases "go-redis/internal/usecases/redis"
	"go-redis/internal/usecases/redis/commands"

	"github.com/labstack/echo/v4"
)

type IRespController interface {
	ParseResp(c echo.Context) error
	ParseAndStoreResp(c echo.Context) error
	GetStoredData(c echo.Context) error
	DeleteStoredData(c echo.Context) error
	GetCommandHistory(c echo.Context) error
}

type RespController struct {
	respUseCases redis_usecases.IRedisUseCases
	mu           sync.RWMutex
	requestCount int64
}

func NewRespController(respUseCases redis_usecases.IRedisUseCases) IRespController {
	return &RespController{
		respUseCases: respUseCases,
	}
}

type ParseRespRequest struct {
	RespData string `json:"resp_data" validate:"required"`
}

type ParseRespResponse struct {
	Status string                 `json:"status"`
	Data   map[string]interface{} `json:"data,omitempty"`
	Error  string                 `json:"error,omitempty"`
}

type ParseAndStoreRespResponse struct {
	Status string                        `json:"status"`
	Data   *commands.ParseAndStoreResult `json:"data,omitempty"`
	Error  string                        `json:"error,omitempty"`
}

type GetDataRequest struct {
	KeyName string `json:"key_name" validate:"required"`
}

type GetDataResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

type DeleteDataRequest struct {
	KeyName string `json:"key_name" validate:"required"`
}

type CommandHistoryResponse struct {
	Status   string                `json:"status"`
	Commands []sqlite.RedisCommand `json:"commands,omitempty"`
	Error    string                `json:"error,omitempty"`
}

func (r *RespController) ParseResp(c echo.Context) error {
	r.incrementRequestCount()
	defer r.decrementRequestCount()

	var req ParseRespRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ParseRespResponse{
			Status: "error",
			Error:  "Invalid request format",
		})
	}

	result, err := r.respUseCases.GetCommands().ParseRespToKeyValue.Handle(req.RespData)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ParseRespResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	return c.JSON(http.StatusOK, ParseRespResponse{
		Status: "success",
		Data:   result,
	})
}

func (r *RespController) ParseAndStoreResp(c echo.Context) error {
	r.incrementRequestCount()
	defer r.decrementRequestCount()

	var req ParseRespRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ParseAndStoreRespResponse{
			Status: "error",
			Error:  "Invalid request format",
		})
	}

	result, err := r.respUseCases.GetCommands().ParseAndStoreRespData.Handle(req.RespData)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ParseAndStoreRespResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	return c.JSON(http.StatusOK, ParseAndStoreRespResponse{
		Status: "success",
		Data:   result,
	})
}

func (r *RespController) GetStoredData(c echo.Context) error {
	keyName := c.Param("key")
	if keyName == "" {
		return c.JSON(http.StatusBadRequest, GetDataResponse{
			Status: "error",
			Error:  "Key name is required",
		})
	}

	data, err := r.respUseCases.GetQueries().GetStoredData.Handle(keyName)
	if err != nil {
		return c.JSON(http.StatusNotFound, GetDataResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	return c.JSON(http.StatusOK, GetDataResponse{
		Status: "success",
		Data:   data,
	})
}

func (r *RespController) DeleteStoredData(c echo.Context) error {
	keyName := c.Param("key")
	if keyName == "" {
		return c.JSON(http.StatusBadRequest, GetDataResponse{
			Status: "error",
			Error:  "Key name is required",
		})
	}

	err := r.respUseCases.GetCommands().DeleteStoredData.Handle(keyName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, GetDataResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	return c.JSON(http.StatusOK, GetDataResponse{
		Status: "success",
		Data:   "Key deleted successfully",
	})
}

func (r *RespController) GetCommandHistory(c echo.Context) error {
	limit := 50 // Default limit
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	commands, err := r.respUseCases.GetQueries().GetCommandHistory.Handle(limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, CommandHistoryResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	return c.JSON(http.StatusOK, CommandHistoryResponse{
		Status:   "success",
		Commands: commands,
	})
}

func (r *RespController) incrementRequestCount() {
	r.mu.Lock()
	r.requestCount++
	r.mu.Unlock()
}

func (r *RespController) decrementRequestCount() {
	r.mu.Lock()
	r.requestCount--
	r.mu.Unlock()
}

func (r *RespController) GetActiveRequestCount() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.requestCount
}
