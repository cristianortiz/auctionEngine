package websocket

import (
	"time"

	"github.com/cristianortiz/auctionEngine/internal/shared/logger"
	"github.com/gofiber/websocket/v2"
	"go.uber.org/zap"
)

var log = logger.GetLogger()

// Hub mantiene el registro de clientes y maneja la transmisión de mensajes.
type Hub struct {
	// Registered clients, grouped by lot ID.
	// The keys of the outer map are lot IDs.
	// The inner map keys are clients, and the boolean value is ignored.
	clients map[string]map[*Client]bool
	// Inbound messages from the clients.
	broadcast chan *Message
	// Register requests from the clients.
	register chan *Client
	// Unregister requests from clients.
	unregister chan *Client
}

// Client representa una conexión WebSocket individual.
type Client struct {
	hub *Hub
	// The websocket connection.
	conn *websocket.Conn
	// Buffered channel of outbound messages.
	send chan []byte
	// The lot ID this client is connected to.
	lotID string
}

// Message representa un mensaje para ser transmitido.
type Message struct {
	LotID string
	Data  []byte
}

// NewHub crea una nueva instancia del Hub.
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]map[*Client]bool),
	}
}

// Run inicia el Hub, escuchando en sus canales.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// Registra el cliente en el grupo del lotID
			if _, ok := h.clients[client.lotID]; !ok {
				h.clients[client.lotID] = make(map[*Client]bool)
			}
			h.clients[client.lotID][client] = true
			log.Info("Client registered", zap.String("lotID", client.lotID), zap.String("remote_addr", client.conn.RemoteAddr().String()))

		case client := <-h.unregister:
			// Elimina el cliente del grupo del lotID
			if clients, ok := h.clients[client.lotID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					log.Info("Client unregistered", zap.String("lotID", client.lotID), zap.String("remote_addr", client.conn.RemoteAddr().String()))
					// Si no quedan clientes en este grupo, elimina el mapa
					if len(clients) == 0 {
						delete(h.clients, client.lotID)
						log.Info("Lot group removed as empty", zap.String("lotID", client.lotID))
					}
				}
			}

		case message := <-h.broadcast:
			// Transmite el mensaje a todos los clientes en el grupo del lotID
			if clients, ok := h.clients[message.LotID]; ok {
				log.Debug("Broadcasting message to lot", zap.String("lotID", message.LotID), zap.Int("clients", len(clients)))
				for client := range clients {
					select {
					case client.send <- message.Data:
						// Mensaje enviado
					default:
						// No se pudo enviar, cliente probablemente desconectado
						close(client.send)
						delete(clients, client)
						log.Warn("Failed to send message to client, unregistering", zap.String("lotID", client.lotID), zap.String("remote_addr", client.conn.RemoteAddr().String()))
					}
				}
			}
		}
	}
}

// RegisterClient registra un nuevo cliente en el Hub.
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient elimina un cliente del Hub.
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// BroadcastMessageToLot envía un mensaje a todos los clientes suscritos a un lotID específico.
func (h *Hub) BroadcastMessageToLot(lotID string, data []byte) {
	h.broadcast <- &Message{LotID: lotID, Data: data}
}

// ReadPump lee mensajes del cliente WebSocket y los envía al Hub (a través del canal broadcast).
// Este método debe ejecutarse en una goroutine por cada cliente.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.UnregisterClient(c)
		c.conn.Close()
	}()
	// Configura timeouts si es necesario
	// c.conn.SetReadDeadline(time.Now().Add(pongWait))
	// c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		// Lee el mensaje del cliente
		mt, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Error("WebSocket read error", zap.Error(err), zap.String("lotID", c.lotID), zap.String("remote_addr", c.conn.RemoteAddr().String()))
			}
			break
		}
		// Aquí, en un hub genérico, podrías simplemente reenviar el mensaje
		// o pasarlo a un canal de procesamiento si el hub tuviera esa responsabilidad.
		// Como este hub es *sin lógica de negocio*, no procesamos el mensaje aquí.
		// La lógica de negocio que recibe mensajes del cliente (ej: una puja)
		// se manejará en el módulo 'auction', que recibirá el mensaje
		// a través de un mecanismo que definiremos luego (ej: un canal o callback).

		// Por ahora, solo loggeamos que se recibió un mensaje (opcional)
		log.Debug("Received message from client", zap.String("lotID", c.lotID), zap.String("remote_addr", c.conn.RemoteAddr().String()), zap.Int("message_type", mt), zap.ByteString("message_data", message))

		// Si quisieras que el hub reenviara mensajes recibidos a todos en el lote (no es el caso aquí):
		// c.hub.BroadcastMessageToLot(c.lotID, message)
	}
}

// WritePump envía mensajes del Hub al cliente WebSocket.
// Este método debe ejecutarse en una goroutine por cada cliente.
func (c *Client) WritePump() {
	// Configura un ticker para enviar pings periódicos si es necesario
	// ticker := time.NewTicker(pingPeriod)
	defer func() {
		// ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			// Recibe mensaje del canal de envío del cliente
			if !ok {
				// El Hub cerró el canal.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Escribe el mensaje al cliente
			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Error("WebSocket write error", zap.Error(err), zap.String("lotID", c.lotID), zap.String("remote_addr", c.conn.RemoteAddr().String()))
				return // Sale del loop, defer cerrará la conexión
			}

			// case <-ticker.C:
			// 	// Envía un ping periódico
			// 	if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			// 		return // Sale del loop, defer cerrará la conexión
			// 	}
		}
	}
}

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
