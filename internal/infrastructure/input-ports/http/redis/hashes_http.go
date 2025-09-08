package redis

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HMSET: body: {"key":"k","fields":{"f1":"v1","f2":"v2"}}
func (r *RespController) HMSetHTTP(c echo.Context) error {
	type reqBody struct {
		Key    string            `json:"key"`
		Fields map[string]string `json:"fields"`
	}
	var req reqBody
	if err := c.Bind(&req); err != nil || req.Key == "" || len(req.Fields) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "invalid payload"})
	}
	if err := r.respUseCases.GetCommands().HMSet.Handle(req.Key, req.Fields); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"status": "success"})
}

// HMGET: body: {"key":"k","fields":["f1","f2"]}
func (r *RespController) HMGetHTTP(c echo.Context) error {
	type reqBody struct {
		Key    string   `json:"key"`
		Fields []string `json:"fields"`
	}
	var req reqBody
	if err := c.Bind(&req); err != nil || req.Key == "" || len(req.Fields) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "invalid payload"})
	}
	vals, err := r.respUseCases.GetQueries().HMGet.Handle(req.Key, req.Fields)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"status": "success", "values": vals})
}

