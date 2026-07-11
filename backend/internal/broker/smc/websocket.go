package smc

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

type TickHandler func(message []byte)
type WebSocketErrorHandler func(error)

type WebSocketClient struct {
	client       *Client
	url          string
	reconnectMin time.Duration
	reconnectMax time.Duration

	mu     sync.Mutex
	conn   net.Conn
	closed bool
}

func (c *Client) NewWebSocketClient(webSocketURL string) *WebSocketClient {
	if webSocketURL == "" {
		webSocketURL = c.webSocketURL
	}

	return &WebSocketClient{
		client:       c,
		url:          webSocketURL,
		reconnectMin: time.Second,
		reconnectMax: 30 * time.Second,
	}
}

func (w *WebSocketClient) Run(ctx context.Context, onMessage TickHandler, onError WebSocketErrorHandler) error {
	backoff := w.reconnectMin

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := w.connect(ctx); err != nil {
			if onError != nil {
				onError(err)
			}
			if err := sleep(ctx, backoff); err != nil {
				return err
			}
			backoff = nextBackoff(backoff, w.reconnectMax)
			continue
		}

		backoff = w.reconnectMin
		err := w.readLoop(ctx, onMessage)
		w.closeConn()
		if w.isClosed() {
			return nil
		}

		if err != nil && onError != nil {
			onError(err)
		}
	}
}

func (w *WebSocketClient) Subscribe(ctx context.Context, payload []byte) error {
	w.mu.Lock()
	conn := w.conn
	w.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("websocket is not connected")
	}

	frame, err := encodeClientFrame(1, payload)
	if err != nil {
		return err
	}

	_, err = conn.Write(frame)
	return err
}

func (w *WebSocketClient) Close() error {
	w.mu.Lock()
	w.closed = true
	conn := w.conn
	w.conn = nil
	w.mu.Unlock()

	if conn == nil {
		return nil
	}

	_, _ = conn.Write([]byte{0x88, 0x80, 0, 0, 0, 0})
	return conn.Close()
}

func (w *WebSocketClient) connect(ctx context.Context) error {
	if w.url == "" {
		return fmt.Errorf("websocket url is required")
	}

	parsed, err := url.Parse(w.url)
	if err != nil {
		return fmt.Errorf("parse websocket url: %w", err)
	}

	host := parsed.Host
	if !strings.Contains(host, ":") {
		if parsed.Scheme == "wss" {
			host += ":443"
		} else {
			host += ":80"
		}
	}

	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", host)
	if err != nil {
		return fmt.Errorf("dial websocket: %w", err)
	}

	if parsed.Scheme == "wss" {
		tlsConn := tls.Client(conn, &tls.Config{ServerName: parsed.Hostname(), MinVersion: tls.VersionTLS12})
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			_ = conn.Close()
			return fmt.Errorf("tls websocket handshake: %w", err)
		}
		conn = tlsConn
	}

	key, err := websocketKey()
	if err != nil {
		_ = conn.Close()
		return err
	}

	path := parsed.RequestURI()
	if path == "" {
		path = "/"
	}

	request := "GET " + path + " HTTP/1.1\r\n" +
		"Host: " + parsed.Host + "\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Key: " + key + "\r\n" +
		"Sec-WebSocket-Version: 13\r\n"

	if token := w.client.FeedToken(); token != "" {
		request += "Authorization: Bearer " + token + "\r\n"
	}
	request += "\r\n"

	if _, err := conn.Write([]byte(request)); err != nil {
		_ = conn.Close()
		return fmt.Errorf("write websocket handshake: %w", err)
	}

	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, nil)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("read websocket handshake: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		_ = conn.Close()
		return fmt.Errorf("websocket upgrade failed: status=%d", resp.StatusCode)
	}

	if !validAccept(key, resp.Header.Get("Sec-WebSocket-Accept")) {
		_ = conn.Close()
		return fmt.Errorf("websocket accept validation failed")
	}

	w.mu.Lock()
	w.conn = &bufferedConn{Conn: conn, reader: reader}
	w.mu.Unlock()

	return nil
}

func (w *WebSocketClient) readLoop(ctx context.Context, onMessage TickHandler) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		w.mu.Lock()
		conn := w.conn
		w.mu.Unlock()
		if conn == nil {
			return fmt.Errorf("websocket connection closed")
		}

		opcode, payload, err := readFrame(conn)
		if err != nil {
			return err
		}

		switch opcode {
		case 1, 2:
			if onMessage != nil {
				onMessage(payload)
			}
		case 8:
			return fmt.Errorf("websocket close frame received")
		case 9:
			frame, err := encodeClientFrame(10, payload)
			if err != nil {
				return err
			}
			_, _ = conn.Write(frame)
		}
	}
}

func (w *WebSocketClient) closeConn() {
	w.mu.Lock()
	conn := w.conn
	w.conn = nil
	w.mu.Unlock()

	if conn != nil {
		_ = conn.Close()
	}
}

func (w *WebSocketClient) isClosed() bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.closed
}

func websocketKey() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate websocket key: %w", err)
	}

	return base64.StdEncoding.EncodeToString(raw), nil
}

func validAccept(key string, accept string) bool {
	hash := sha1.Sum([]byte(key + websocketGUID))
	return base64.StdEncoding.EncodeToString(hash[:]) == accept
}

func readFrame(reader io.Reader) (byte, []byte, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(reader, header); err != nil {
		return 0, nil, err
	}

	opcode := header[0] & 0x0f
	masked := header[1]&0x80 != 0
	length := uint64(header[1] & 0x7f)

	switch length {
	case 126:
		extended := make([]byte, 2)
		if _, err := io.ReadFull(reader, extended); err != nil {
			return 0, nil, err
		}
		length = uint64(binary.BigEndian.Uint16(extended))
	case 127:
		extended := make([]byte, 8)
		if _, err := io.ReadFull(reader, extended); err != nil {
			return 0, nil, err
		}
		length = binary.BigEndian.Uint64(extended)
	}

	var maskKey []byte
	if masked {
		maskKey = make([]byte, 4)
		if _, err := io.ReadFull(reader, maskKey); err != nil {
			return 0, nil, err
		}
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return 0, nil, err
	}

	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	return opcode, payload, nil
}

func encodeClientFrame(opcode byte, payload []byte) ([]byte, error) {
	mask := make([]byte, 4)
	if _, err := rand.Read(mask); err != nil {
		return nil, fmt.Errorf("generate websocket mask: %w", err)
	}

	header := []byte{0x80 | opcode}
	length := len(payload)

	switch {
	case length < 126:
		header = append(header, 0x80|byte(length))
	case length <= 65535:
		header = append(header, 0x80|126, byte(length>>8), byte(length))
	default:
		header = append(header, 0x80|127)
		size := make([]byte, 8)
		binary.BigEndian.PutUint64(size, uint64(length))
		header = append(header, size...)
	}

	frame := append(header, mask...)
	for i, b := range payload {
		frame = append(frame, b^mask[i%4])
	}

	return frame, nil
}

func nextBackoff(current time.Duration, max time.Duration) time.Duration {
	next := current * 2
	if next > max {
		return max
	}

	return next
}

type bufferedConn struct {
	net.Conn
	reader *bufio.Reader
}

func (c *bufferedConn) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}
