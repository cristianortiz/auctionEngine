package httpserver

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/cristianortiz/auctionEngine/internal/shared/logger"
	"github.com/cristianortiz/auctionEngine/internal/shared/websocket"
	"github.com/gofiber/fiber/v2"
	fws "github.com/gofiber/websocket/v2" // Alias to avoid name conflicts
	"go.uber.org/zap"
)

type Server struct {
	app *fiber.App
	hub *websocket.Hub // wbs hub reference
}

var log = logger.GetLogger() // logger instance
// NewServer creates a new server instance, receiving wbs hub
func NewServer(addr string, hub *websocket.Hub) *Server {
	app := fiber.New()

	// Middleware for logging
	app.Use(func(c *fiber.Ctx) error {
		log.Info("HTTP request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("remote_addr", c.IP()),
		)
		return c.Next()
	})

	// health check EP
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK, Welcome to AuctionEngine Project")
	})

	//fiber requires the WBS base route, like  /ws, has to managed by a middleware
	app.Use("/ws", func(c *fiber.Ctx) error {
		//returns true if the request is a WBS upgrade
		if fws.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	//defines the specific route for auction by lotID
	app.Get("/ws/auction/:lotid", fws.New(func(c *fws.Conn) {
		//extract lotid parameters from url
		lotID := c.Params("lotid")
		if lotID == "" {
			log.Error("webSocket connection attempt whithout lotID")
			c.Close()
			return
		}
		log.Info("New WebSocket connection attempt", zap.String("lotID", lotID), zap.String("remote_addr", c.RemoteAddr().String()))

		//creates a new client instance
		client := &websocket.Client{
			Hub:   hub, //assigns the hub reference received by the server
			Conn:  c,
			Send:  make(chan []byte, 256),
			LotID: lotID,
		}

		//register the client in the hub
		hub.RegisterClient(client)
		// starts the goroutines to write and red client messages
		go client.WritePump()
		client.ReadPump() //ReadPump blocks, its execute int handler goroutine
		//ReadPump exits when connections closes or there ir an error
		//defer function in ReadPump,takes care of unregister and close the connection

	}))

	srv := &Server{
		app: app,
		hub: hub,
	}

	return srv
}

func (s *Server) Start(addr string) error {
	// Manejo de cierre limpio con señal
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit

		log.Info("Shutting down HTTP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.app.ShutdownWithContext(ctx)
	}()

	log.Info("HTTP server started", zap.String("addr", addr))
	return s.app.Listen(addr)
}
