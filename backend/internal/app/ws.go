package app

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
)

const (
	webSocketGUID      = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	maxClientFrameSize = 1 << 20 // 1MB
)

type wsHub struct {
	mu     sync.RWMutex
	byUser map[string]map[*wsClient]struct{}
}

type wsClient struct {
	userID  string
	conn    net.Conn
	writeMu sync.Mutex
}

func newWSHub() *wsHub {
	return &wsHub{
		byUser: make(map[string]map[*wsClient]struct{}),
	}
}

func (h *wsHub) register(client *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.byUser[client.userID]; !ok {
		h.byUser[client.userID] = make(map[*wsClient]struct{})
	}
	h.byUser[client.userID][client] = struct{}{}
}

func (h *wsHub) unregister(client *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.byUser[client.userID]
	if clients == nil {
		return
	}
	delete(clients, client)
	if len(clients) == 0 {
		delete(h.byUser, client.userID)
	}
}

func (h *wsHub) sendToUser(userID string, message any) {
	h.mu.RLock()
	userClients := h.byUser[userID]
	clients := make([]*wsClient, 0, len(userClients))
	for client := range userClients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		if err := client.sendJSON(message); err != nil {
			h.unregister(client)
			_ = client.conn.Close()
		}
	}
}

func (h *wsHub) sendToUsers(userIDs []string, message any) {
	seen := make(map[string]struct{}, len(userIDs))
	for _, userID := range userIDs {
		trimmed := strings.TrimSpace(userID)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		h.sendToUser(trimmed, message)
	}
}

func (a *App) pushRealtimeEventToUsers(userIDs []string, eventType string, data map[string]any) {
	if strings.TrimSpace(eventType) == "" {
		return
	}

	a.wsHub.sendToUsers(userIDs, map[string]any{
		"type": eventType,
		"data": data,
	})
}

func (a *App) loadAllUserIDs(ctx context.Context) ([]string, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT id
		FROM users
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userIDs := make([]string, 0)
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, rows.Err()
}

func (c *wsClient) sendJSON(message any) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return c.writeFrame(0x1, payload)
}

func (c *wsClient) writeFrame(opcode byte, payload []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	header := make([]byte, 0, 10)
	header = append(header, 0x80|(opcode&0x0F))

	length := len(payload)
	switch {
	case length <= 125:
		header = append(header, byte(length))
	case length <= 65535:
		header = append(header, 126)
		ext := make([]byte, 2)
		binary.BigEndian.PutUint16(ext, uint16(length))
		header = append(header, ext...)
	default:
		header = append(header, 127)
		ext := make([]byte, 8)
		binary.BigEndian.PutUint64(ext, uint64(length))
		header = append(header, ext...)
	}

	if _, err := c.conn.Write(header); err != nil {
		return err
	}
	if length == 0 {
		return nil
	}

	_, err := c.conn.Write(payload)
	return err
}

func (a *App) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	currentUser, err := a.userFromRequest(r.Context(), r)
	if err != nil {
		if isUnauthorizedError(err) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load current user")
		return
	}

	conn, status, err := upgradeToWebSocket(w, r)
	if err != nil {
		if status != 0 {
			writeError(w, status, err.Error())
		}
		return
	}

	client := &wsClient{
		userID: currentUser.ID,
		conn:   conn,
	}
	a.wsHub.register(client)
	defer func() {
		a.wsHub.unregister(client)
		_ = conn.Close()
	}()

	for {
		opcode, payload, err := readClientFrame(conn)
		if err != nil {
			return
		}

		switch opcode {
		case 0x8: // close
			_ = client.writeFrame(0x8, nil)
			return
		case 0x9: // ping
			_ = client.writeFrame(0xA, payload)
		case 0x1, 0x2, 0xA:
			// Text/Binary/Pong payloads from client are ignored.
		default:
			return
		}
	}
}

func upgradeToWebSocket(w http.ResponseWriter, r *http.Request) (net.Conn, int, error) {
	if !headerContainsToken(r.Header.Get("Connection"), "upgrade") || !strings.EqualFold(strings.TrimSpace(r.Header.Get("Upgrade")), "websocket") {
		return nil, http.StatusBadRequest, errors.New("websocket upgrade required")
	}

	if strings.TrimSpace(r.Header.Get("Sec-WebSocket-Version")) != "13" {
		return nil, http.StatusBadRequest, errors.New("unsupported websocket version")
	}

	key := strings.TrimSpace(r.Header.Get("Sec-WebSocket-Key"))
	if key == "" {
		return nil, http.StatusBadRequest, errors.New("missing Sec-WebSocket-Key")
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, http.StatusInternalServerError, errors.New("websocket hijacking not supported")
	}

	conn, _, err := hijacker.Hijack()
	if err != nil {
		return nil, 0, err
	}

	response := fmt.Sprintf(
		"HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: %s\r\n\r\n",
		webSocketAccept(key),
	)
	if _, err := io.WriteString(conn, response); err != nil {
		_ = conn.Close()
		return nil, 0, err
	}

	return conn, 0, nil
}

func readClientFrame(conn net.Conn) (byte, []byte, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return 0, nil, err
	}

	fin := (header[0] & 0x80) != 0
	opcode := header[0] & 0x0F
	if !fin && opcode != 0x0 {
		return 0, nil, errors.New("fragmented frames are not supported")
	}

	masked := (header[1] & 0x80) != 0
	if !masked {
		return 0, nil, errors.New("client frame must be masked")
	}

	payloadLen := uint64(header[1] & 0x7F)
	switch payloadLen {
	case 126:
		ext := make([]byte, 2)
		if _, err := io.ReadFull(conn, ext); err != nil {
			return 0, nil, err
		}
		payloadLen = uint64(binary.BigEndian.Uint16(ext))
	case 127:
		ext := make([]byte, 8)
		if _, err := io.ReadFull(conn, ext); err != nil {
			return 0, nil, err
		}
		payloadLen = binary.BigEndian.Uint64(ext)
	}

	if payloadLen > maxClientFrameSize {
		return 0, nil, errors.New("websocket frame too large")
	}

	maskKey := make([]byte, 4)
	if _, err := io.ReadFull(conn, maskKey); err != nil {
		return 0, nil, err
	}

	payload := make([]byte, payloadLen)
	if payloadLen > 0 {
		if _, err := io.ReadFull(conn, payload); err != nil {
			return 0, nil, err
		}
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	if opcode == 0x0 {
		return 0, nil, errors.New("continuation frames are not supported")
	}

	return opcode, payload, nil
}

func headerContainsToken(value, token string) bool {
	for _, part := range strings.Split(value, ",") {
		if strings.EqualFold(strings.TrimSpace(part), token) {
			return true
		}
	}
	return false
}

func webSocketAccept(key string) string {
	hash := sha1.Sum([]byte(key + webSocketGUID))
	return base64.StdEncoding.EncodeToString(hash[:])
}
