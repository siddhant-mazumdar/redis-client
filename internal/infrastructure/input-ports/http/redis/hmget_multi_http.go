package redis

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HMGETM: body: {"keys": {"k1": ["f1","f2"], "k2": ["f1"]}}
func (r *RespController) HMGetMultiHTTP(c echo.Context) error {
	type reqBody struct {
		Keys map[string][]string `json:"keys"`
	}
	var req reqBody
	if err := c.Bind(&req); err != nil || len(req.Keys) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"status": "error", "error": "invalid payload"})
	}
	res, err := r.respUseCases.GetQueries().HMGetMulti.Handle(req.Keys)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"status": "error", "error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"status": "success", "data": res})
}
