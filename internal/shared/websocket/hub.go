package websocket

import (
	"context"
	"time"

	"github.com/cristianortiz/auctionEngine/internal/shared/logger"
	"github.com/gofiber/websocket/v2"
	"go.uber.org/zap"
)

var log = logger.GetLogger()

// Constants for WebSocket configuration (adjust as needed)
const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Hub keeps client's registry and handle messages broadcasting
type Hub struct {
	// Registered clients, grouped by lot ID.
	// The keys of the outer map are lot IDs.
	// The inner map keys are clients, and the boolean value is ignored.
	clients map[string]map[*Client]bool
	// Inbound messages from the clien
	broadcast chan *Message
	// Register requests from the clients.
	register chan *Client
	// Unregister requests from clients.
	unregister      chan *Client
	InboundMessages chan *ClientMessage // this channel will be listened to by module-specific handlers (e.g, auction handler)
}

// Client represents a ws individual connection
type Client struct {
	Hub *Hub
	// The websocket connection.
	Conn *websocket.Conn
	// Buffered channel of outbound messages.
	Send chan []byte
	// The lot ID this client is connected to.
	LotID string
	// Unique identifier for the client
	ID string
}

type Message struct {
	LotID string
	Data  []byte
}

// ClientMessage is used for wraping the client and data message received.
// is used to send inbound messages from the client to the hub handlers
type ClientMessage struct {
	Client *Client
	Data   []byte
}

func NewHub() *Hub {
	return &Hub{
		broadcast:       make(chan *Message),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		clients:         make(map[string]map[*Client]bool),
		InboundMessages: make(chan *ClientMessage),
	}
}

// Run starts the hub listening in their channels
func (h *Hub) Run(ctx context.Context) {
	log.Info("Websocker Hub started")
	for {
		select {
		case <-ctx.Done(): // <-- Check context cancellation
			log.Info("WebSocket Hub shutting down due to context cancellation")
			// TODO: Consider graceful shutdown of clients
			return // Exit the goroutine
		case client := <-h.register:
			// Register the client in lotId group
			if _, ok := h.clients[client.LotID]; !ok {
				h.clients[client.LotID] = make(map[*Client]bool)
			}
			h.clients[client.LotID][client] = true
			log.Info("Client registered",
				zap.String("clientID", client.ID),
				zap.String("LotID", client.LotID),
				zap.String("remote_addr", client.Conn.RemoteAddr().String()),
				zap.Int("total_clients", func() int {
					count := 0
					for _, lotClients := range h.clients {
						count += len(lotClients)
					}
					return count
				}()),
			)

		case client := <-h.unregister:
			// remove the client from LotID group
			if clients, ok := h.clients[client.LotID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					log.Info("Client unregistered",
						zap.String("clientID", client.ID),
						zap.String("lotID", client.LotID),
						zap.String("remote_addr", client.Conn.RemoteAddr().String()),
						zap.Int("total_clients", func() int { // Log total clients
							count := 0
							for _, lotClients := range h.clients {
								count += len(lotClients)
							}
							return count
						}()),
					)
					// Si no quedan clientes en este grupo, elimina el mapa
					if len(clients) == 0 {
						delete(h.clients, client.LotID)
						log.Info("Lot group removed as empty", zap.String("LotID", client.LotID))
					}
				}
			}

		case message := <-h.broadcast:
			//broadcast the message to all the clients in LotID group
			if clients, ok := h.clients[message.LotID]; ok {
				log.Debug("Broadcasting message to lot", zap.String("LotID", message.LotID), zap.Int("clients", len(clients)))
				for client := range clients {
					select {
					case client.Send <- message.Data:
						// message sended
					default:
						//message could not be sent, client probably disconneted, closing channel
						close(client.Send)
						//deleting client form client's map
						delete(clients, client)
						log.Warn("Failed to Send message to client, unregistering",
							zap.String("clientID", client.ID), // Use client.ID
							zap.String("lotID", client.LotID),
							zap.String("remote_addr", client.Conn.RemoteAddr().String()),
						)
					}
				}
			}
		}
	}
}

// RegisterClient register a new client in the hub
func (h *Hub) RegisterClient(client *Client) {
	select { // Use select to avoid blocking if channel is full
	case h.register <- client:
		log.Debug("Client queued for registration",
			zap.String("clientID", client.ID),
			zap.String("lotID", client.LotID),
		)
	default:
		log.Error("Register channel is full, client registration failed",
			zap.String("clientID", client.ID),
			zap.String("lotID", client.LotID),
		)
		// Optionally close the client connection immediately if registration fails
		_ = client.Conn.Close()
	}
}

// UnregisterClient delete a client from the hub
func (h *Hub) UnregisterClient(client *Client) {
	select { // Use select to avoid blocking if channel is full
	case h.unregister <- client:
		log.Debug("Client queued for unregistration",
			zap.String("clientID", client.ID),
			zap.String("lotID", client.LotID),
		)
	default:
		log.Error("Unregister channel is full, client unregistration failed",
			zap.String("clientID", client.ID),
			zap.String("lotID", client.LotID),
		)
		// The client might already be closing, not much to do here.
	}
}

// BroadcastMessageToLot envía un mensaje a todos los clientes suscritos a un lotID específico.
func (h *Hub) BroadcastMessageToLot(lotID string, data []byte) {
	select { // Use select to avoid blocking if channel is full
	case h.broadcast <- &Message{LotID: lotID, Data: data}:
		log.Debug("Message queued for broadcast", zap.String("lotID", lotID))
	default:
		log.Error("Broadcast channel is full, message dropped", zap.String("lotID", lotID))
		// Handle case where broadcast channel is full (e.g., log error, implement retry)
	}
}

// ReadPump lee mensajes del cliente WebSocket y los envía al Hub (a través del canal broadcast).
// Este método debe ejecutarse en una goroutine por cada cliente.
func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		c.Hub.UnregisterClient(c)
		c.Conn.Close()
		log.Info("ReadPump stopped for client",
			zap.String("clientID", c.ID),
			zap.String("lotID", c.LotID),
			zap.String("remote_addr", c.Conn.RemoteAddr().String()),
		)
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	log.Info("ReadPump started for client",
		zap.String("clientID", c.ID),
		zap.String("lotID", c.LotID),
		zap.String("remote_addr", c.Conn.RemoteAddr().String()),
	)

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			log.Info("ReadPump context cancelled for client",
				zap.String("clientID", c.ID), // Use client.ID
				zap.String("lotID", c.LotID),
			)
			return // Exit the goroutine
		default:
			// Continue reading
		}

		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Error("WebSocket read error",
					zap.String("clientID", c.ID), // Use client.ID
					zap.String("lotID", c.LotID),
					zap.String("remote_addr", c.Conn.RemoteAddr().String()),
					zap.Error(err),
				)
			} else {
				log.Info("WebSocket connection closed by peer",
					zap.String("clientID", c.ID), // Use client.ID
					zap.String("lotID", c.LotID),
					zap.String("remote_addr", c.Conn.RemoteAddr().String()),
					zap.Error(err), // Log the specific close error
				)
			}
			break // Exit the loop on read error
		}
		// message = bytes.TrimSpace(bytes.ReplaceAll(message, newline, space)) // Optional: trim whitespace

		log.Debug("Received message from client",
			zap.String("clientID", c.ID), // Use client.ID
			zap.String("lotID", c.LotID),
			zap.ByteString("message", message),
		)

		// Send the received message to the Hub's InboundMessages channel
		// Module-specific handlers will listen on this channel.
		select {
		case c.Hub.InboundMessages <- &ClientMessage{Client: c, Data: message}: // <-- Send message to InboundMessages
			log.Debug("Message sent to Hub's InboundMessages channel",
				zap.String("clientID", c.ID), // Use client.ID
				zap.String("lotID", c.LotID),
			)
		default:
			// If the inbound channel is full, it means handlers are not keeping up.
			// Log an error or implement backpressure/dropping logic.
			log.Error("Hub InboundMessages channel is full, dropping message",
				zap.String("clientID", c.ID), // Use client.ID
				zap.String("lotID", c.LotID),
				zap.ByteString("message", message),
			)
			// Optionally send an error back to the client? Might be too late.
		}
	}

}

// / WritePump pumps messages from the hub to the websocket connection.
// A goroutine running WritePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// invoking WriteControl and WriteMessage from a single goroutine.
func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Hub.UnregisterClient(c)
		c.Conn.Close()
		log.Info("WritePump stopped for client",
			zap.String("clientID", c.ID),
			zap.String("lotID", c.LotID),
			zap.String("remote_addr", c.Conn.RemoteAddr().String()),
		)
	}()

	log.Info("WritePump started for client",
		zap.String("clientID", c.ID),
		zap.String("lotID", c.LotID),
		zap.String("remote_addr", c.Conn.RemoteAddr().String()),
	)

	for {
		select {
		case <-ctx.Done():
			log.Info("WritePump context cancelled for client",
				zap.String("clientID", c.ID),
				zap.String("lotID", c.LotID),
			)
			// Attempt to send a close message before exiting
			err := c.Conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(writeWait))
			if err != nil {
				log.Error("Failed to send close control message",
					zap.String("clientID", c.ID),
					zap.String("lotID", c.LotID),
					zap.Error(err),
				)
			}
			return // Exit the goroutine

		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The Hub closed the channel.
				log.Info("Client send channel closed by Hub",
					zap.String("clientID", c.ID),
					zap.String("lotID", c.LotID),
				)
				err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Error("Failed to write close message after channel close",
						zap.String("clientID", c.ID),
						zap.String("lotID", c.LotID),
						zap.Error(err),
					)
				}
				return // Exit the goroutine
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Error("Failed to get next writer for client",
					zap.String("clientID", c.ID),
					zap.String("lotID", c.LotID),
					zap.Error(err),
				)
				return // Exit the goroutine on writer error
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			// This part might need adjustment depending on your message queuing strategy.
			// If you send one message at a time, this loop might not be needed.
			// If you batch messages, ensure they are properly delimited (like with newline).
			n := len(c.Send)
			for range n {
				w.Write([]byte{'\n'}) // Use newline constant if defined
				msg, ok := <-c.Send
				if !ok {
					// Channel closed while draining
					log.Warn("Client send channel closed while draining",
						zap.String("clientID", c.ID),
						zap.String("lotID", c.LotID),
					)
					break
				}
				w.Write(msg)
			}

			if err := w.Close(); err != nil {
				log.Error("Failed to close writer for client",
					zap.String("clientID", c.ID),
					zap.String("lotID", c.LotID),
					zap.Error(err),
				)
				return // Exit the goroutine on close error
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
				log.Error("Failed to write ping message to client",
					zap.String("clientID", c.ID),
					zap.String("lotID", c.LotID),
					zap.Error(err),
				)
				return // Exit the goroutine on ping error
			}
		}
	}
}
