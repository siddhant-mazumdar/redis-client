package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

var SERVICE_ENDPOINT string

func (s *server) initializeRoutes() {
	SERVICE_ENDPOINT = s.config.GetConfig().ServiceEndpointPrefix
	addSampleRoutes(s)
	addRedisRoutes(s)
}

func addSampleRoutes(s *server) {
	s.echo.GET("/health-check", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"Status": "success",
		})
	})
	s.echo.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"Status": "success",
		})
	})
}

func addRedisRoutes(s *server) {
	s.echo.POST(SERVICE_ENDPOINT+"/v1/redis/parse-resp", s.controllers.RespController.ParseResp)
	s.echo.POST(SERVICE_ENDPOINT+"/v1/redis/parse-and-store", s.controllers.RespController.ParseAndStoreResp)
	s.echo.POST(SERVICE_ENDPOINT+"/v1/redis/get-data", s.controllers.RespController.GetStoredData)
	s.echo.POST(SERVICE_ENDPOINT+"/v1/redis/delete-data", s.controllers.RespController.DeleteStoredData)
	s.echo.GET(SERVICE_ENDPOINT+"/v1/redis/command-history", s.controllers.RespController.GetCommandHistory)
}
