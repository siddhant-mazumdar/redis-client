package http

import (
	"context"
	"go-redis/config"
	"go-redis/internal/usecases"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type server struct {
	useCases    usecases.IUseCases
	echo        *echo.Echo
	controllers *Controllers
	config      config.IConfig
	messageBus  *MessageBus
	mu          sync.RWMutex
	shutdown    chan os.Signal
}

func NewServer(useCases usecases.IUseCases, config config.IConfig) *server {
	server := &server{
		useCases:    useCases,
		echo:        echo.New(),
		controllers: NewControllers(useCases),
		config:      config,
		messageBus:  NewMessageBus(10), // 10 worker goroutines
		shutdown:    make(chan os.Signal, 1),
	}

	// Configure Echo to handle requests asynchronously
	server.echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: middleware.DefaultSkipper,
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
		CustomTimeFormat: "2006-01-02 15:04:05.00000",
	}))

	// Remove this line to fix the issue:
	// server.echo.Use(server.asyncRequestMiddleware)

	server.addHealthCheckRoutes()
	server.initializeRoutes()

	// Start message bus
	server.messageBus.Start()

	return server
}

func (s *server) asyncRequestMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Create a unique request ID
		requestID := c.Response().Header().Get(echo.HeaderXRequestID)
		if requestID == "" {
			requestID = strconv.FormatInt(time.Now().UnixNano(), 36)
		}

		// Create response channel
		responseChan := make(chan interface{}, 1)

		// Create message for the request
		msg := Message{
			ID:        requestID,
			Type:      RequestMessage,
			Data:      map[string]interface{}{"context": c, "handler": next},
			Timestamp: time.Now(),
			Context:   c.Request().Context(),
			Response:  responseChan,
		}

		// Send message to message bus
		s.messageBus.SendRequest(msg)

		// Wait for response or timeout
		select {
		case response := <-responseChan:
			// Handle the response from message bus
			if err, ok := response.(error); ok {
				return err
			}
			return nil
		case <-time.After(30 * time.Second):
			return echo.NewHTTPError(408, "Request timeout")
		case <-c.Request().Context().Done():
			return echo.NewHTTPError(499, "Client disconnected")
		}
	}
}

func (s *server) addHealthCheckRoutes() {
	s.echo.GET("/health-check", func(c echo.Context) error {
		return c.String(200, "OK")
	})
	s.echo.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})
	s.echo.GET("/", func(c echo.Context) error {
		return c.String(200, "Welcome to redis server")
	})
}

func (s *server) Start() {
	port := 10001
	if os.Getenv("PORT") != "" {
		if portNo, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
			port = portNo
		}
	}

	// Setup graceful shutdown
	signal.Notify(s.shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %d", port)
		if err := s.echo.Start(":" + strconv.Itoa(port)); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Run on main thread - wait for shutdown signal
	s.runMainLoop()
}

func (s *server) runMainLoop() {
	log.Println("Server running on main thread, waiting for requests...")

	// Main thread loop - handle shutdown and monitoring
	for {
		select {
		case sig := <-s.shutdown:
			log.Printf("Received signal: %v", sig)
			s.gracefulShutdown()
			return
		case <-time.After(10 * time.Second):
			// Periodic health check or maintenance tasks
			s.performMaintenanceTasks()
		}
	}
}

func (s *server) performMaintenanceTasks() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Log active connections, cleanup, etc.
	log.Printf("Server health check - Time: %v", time.Now())
}

func (s *server) gracefulShutdown() {
	log.Println("Starting graceful shutdown...")

	// Stop message bus
	s.messageBus.Stop()

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown Echo server
	if err := s.echo.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("Server shutdown complete")
}
