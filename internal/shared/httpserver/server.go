package httpserver

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/cristianortiz/auctionEngine/internal/shared/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Server struct {
	app *fiber.App
}

var log = logger.GetLogger() // Instancia logger para el pakg

func NewServer() *Server {
	app := fiber.New()

	// Middleware de logging
	app.Use(func(c *fiber.Ctx) error {
		log.Info("HTTP request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("remote_addr", c.IP()),
		)
		return c.Next()
	})

	// Endpoint de health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	return &Server{app: app}
}

func (s *Server) Start(addr string) error {
	// Manejo de cierre limpio con señal
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit

		logger.GetLogger().Info("Shutting down HTTP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.app.ShutdownWithContext(ctx)
	}()

	logger.GetLogger().Info("HTTP server started", zap.String("addr", addr))
	return s.app.Listen(addr)
}
