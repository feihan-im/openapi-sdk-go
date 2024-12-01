package fhcore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/feihan-im/openapi-sdk-go/internal/model"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
)

type EventHandler func(ctx context.Context, header *EventHeader, body []byte) error

type EventHeader struct {
	EventId        string `json:"event_id,omitempty"`
	EventType      string `json:"event_type,omitempty"`
	EventCreatedAt uint64 `json:"event_created_at,omitempty"`
}

type defaultWsClient struct {
	config             *Config
	secret             string
	getToken           func(ctx context.Context) (string, error)
	ensurePing         func(ctx context.Context) error
	initOnce           sync.Once
	eventHandlerMap    sync.Map
	eventUnsupportType sync.Map

	reconnectCheckInterval time.Duration
	healthCheckInterval    time.Duration
	aliveTimeout           time.Duration
	connectTimeout         time.Duration
	writeTimeout           time.Duration

	socket         *websocket.Conn
	isReconnecting bool
	isConnecting   bool
	shouldClose    bool
	mu             sync.Mutex

	reqCount       int
	reqCallbacks   map[string]*defaultWsReqCallbacks
	reqCallbacksMu sync.Mutex

	reconnectAttempt     int
	lastMessageAt        int64
	healthCheckTicker    *time.Ticker
	reconnectCheckTicker *time.Ticker
	reconnectTimer       *time.Timer
}

type defaultWsReqCallbacks struct {
	resp chan *model.HttpResponse
}

type defaultWsEventHandlerSet struct {
	mu       sync.RWMutex
	handlers []EventHandler
}

type defaultWsClientOptions struct {
	config     *Config
	secret     string
	getToken   func(ctx context.Context) (string, error)
	ensurePing func(ctx context.Context) error
}

func newDefaultWsClient(options *defaultWsClientOptions) *defaultWsClient {
	return &defaultWsClient{
		config:     options.config,
		secret:     options.secret,
		getToken:   options.getToken,
		ensurePing: options.ensurePing,

		reconnectCheckInterval: 10 * time.Second,
		healthCheckInterval:    20 * time.Second,
		aliveTimeout:           40 * time.Second,
		connectTimeout:         5 * time.Second,
		writeTimeout:           60 * time.Second,
	}
}

func (c *defaultWsClient) init(ctx context.Context) {
	c.ensurePing(ctx)
	_ = c.connect(ctx)
}

func (c *defaultWsClient) OnEvent(eventType string, handler EventHandler) {
	v, ok := c.eventHandlerMap.Load(eventType)
	if !ok {
		v, _ = c.eventHandlerMap.LoadOrStore(eventType, &defaultWsEventHandlerSet{})
	}
	set := v.(*defaultWsEventHandlerSet)
	set.mu.Lock()
	set.handlers = append(set.handlers, handler)
	set.mu.Unlock()
}

func (c *defaultWsClient) OffEvent(eventType string, handler EventHandler) {
	v, ok := c.eventHandlerMap.Load(eventType)
	if !ok {
		return
	}
	hp := reflect.ValueOf(handler).Pointer()
	set := v.(*defaultWsEventHandlerSet)
	set.mu.Lock()
	i := 0
	for ; i < len(set.handlers); i++ {
		p := reflect.ValueOf(set.handlers[i]).Pointer()
		if hp == p {
			break
		}
	}
	if i < len(set.handlers) {
		for ; i+1 < len(set.handlers); i++ {
			set.handlers[i] = set.handlers[i+1]
		}
		set.handlers = set.handlers[0 : len(set.handlers)-1]
	}
	set.mu.Unlock()
}

func (c *defaultWsClient) HttpRequest(ctx context.Context, req *model.HttpRequest) (*model.HttpResponse, error) {
	c.initOnce.Do(func() {
		c.init(ctx)
	})

	req.ReqId = c.newReqId()
	respCh := make(chan *model.HttpResponse)

	c.reqCallbacksMu.Lock()
	c.reqCallbacks[req.ReqId] = &defaultWsReqCallbacks{
		resp: respCh,
	}
	c.reqCallbacksMu.Unlock()

	timer := time.AfterFunc(c.config.RequestTimeout, func() {
		close(respCh)
	})

	err := c.sendMessage(ctx, &model.WebSocketMessage{
		Content: &model.WebSocketMessage_HttpRequest{
			HttpRequest: req,
		},
	})
	if err != nil {
		c.reqCallbacksMu.Lock()
		delete(c.reqCallbacks, req.ReqId)
		c.reqCallbacksMu.Unlock()
		timer.Stop()
		return nil, err
	}

	resp, ok := <-respCh
	timer.Stop()
	if !ok {
		return nil, &ApiError{Code: -1, Msg: "Request timeout"}
	}

	return resp, nil
}

func (c *defaultWsClient) connect(ctx context.Context) (err error) {
	c.mu.Lock()
	err = c.connectUnsafe(ctx)
	c.mu.Unlock()
	return
}

func (c *defaultWsClient) connectUnsafe(ctx context.Context) (err error) {
	if len(c.config.BackendUrl) < 4 {
		return errors.New("invalid backend_url")
	}
	if c.isConnecting {
		return nil
	}
	for _, cb := range c.reqCallbacks {
		close(cb.resp)
	}
	c.reqCount = 0
	c.reqCallbacks = map[string]*defaultWsReqCallbacks{}
	c.lastMessageAt = 0
	c.clearTimerUnsafe()

	c.ensurePing(ctx)
	token, err := c.getToken(ctx)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("ws%s?token=%s", c.config.BackendUrl[4:]+defaultWsPath, token)

	c.config.Logger.Infof(ctx, "[WebSocket] Connecting")

	socket, _, e := websocket.DefaultDialer.Dial(url, nil)
	if e != nil {
		return e
	}
	c.socket = socket

	c.config.Logger.Infof(ctx, "[WebSocket] Connected")

	defer func() {
		if err != nil {
			c.config.Logger.Errorf(ctx, "[WebSocket] Error: %s", err.Error())
			socket.Close()
			c.socket = nil
			c.isConnecting = false
		}
	}()

	socket.SetCloseHandler(func(code int, text string) error {
		c.mu.Lock()
		c.config.Logger.Infof(ctx, "[WebSocket] Closed: code=%v, text=%v", code, text)
		c.socket = nil
		if c.shouldClose {
			c.shouldClose = false
		} else {
			c.reconnectUnsafe(ctx, false)
		}
		c.mu.Unlock()

		return nil
	})

	// send init request
	{
		body, _ := proto.Marshal(&model.WebSocketMessage{
			Content: &model.WebSocketMessage_InitRequest_{
				InitRequest: &model.WebSocketMessage_InitRequest{
					UserAgent: UserAgent,
				},
			},
		})
		secureMessage, err := encryptMessage(c.secret, body)
		if err != nil {
			return err
		}
		data, _ := proto.Marshal(secureMessage)

		if err := socket.WriteMessage(websocket.BinaryMessage, data); err != nil {
			return err
		}
	}

	// receive init response
	{
		_, data, err := socket.ReadMessage()
		if err != nil {
			return err
		}
		c.lastMessageAt = getSystemTimestamp()

		var secureMessage model.SecureMessage
		err = proto.Unmarshal(data, &secureMessage)
		if err != nil {
			return err
		}
		body, err := decryptMessage(c.secret, &secureMessage)
		if err != nil {
			return err
		}
		var message model.WebSocketMessage
		err = proto.Unmarshal(body, &message)
		if err != nil {
			return err
		}

		switch message.Content.(type) {
		case *model.WebSocketMessage_InitResponse_:
			break
		default:
			return errors.New("invalid init response")
		}
	}

	c.config.Logger.Infof(ctx, "[WebSocket] Established")

	go c.handleMessage(ctx, socket)

	c.isConnecting = false
	c.lastMessageAt = getSystemTimestamp()
	c.startHealthCheckUnsafe(ctx)
	c.startReconnectCheckUnsafe(ctx)
	if c.shouldClose {
		c.shouldClose = false
		c.closeUnsafe()
	}

	return nil
}

func (c *defaultWsClient) Close() (err error) {
	c.mu.Lock()
	err = c.closeUnsafe()
	c.mu.Unlock()
	return
}

func (c *defaultWsClient) closeUnsafe() (err error) {
	if c.isConnecting {
		c.shouldClose = true
		return nil
	}
	if c.isReconnecting {
		if c.reconnectTimer != nil {
			c.reconnectTimer.Stop()
		}
		c.reconnectAttempt = 0
		c.reconnectTimer = nil
		return nil
	}
	s := c.socket
	c.socket = nil
	c.shouldClose = true
	if s != nil {
		err = s.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Now().Add(c.writeTimeout),
		)
		if err != nil {
			return err
		}
		err = s.Close()
		if err != nil {
			return err
		}
	}
	c.clearTimerUnsafe()
	return nil
}

func (c *defaultWsClient) reconnectUnsafe(ctx context.Context, force bool) {
	if c.isConnecting {
		return
	}
	if c.isReconnecting && !force {
		return
	}
	if c.shouldClose {
		c.shouldClose = false
	}
	if force {
		if c.reconnectTimer != nil {
			c.reconnectTimer.Stop()
		}
		c.reconnectAttempt = 0
	}
	delay := c.getReconnectDelay()
	c.config.Logger.Infof(ctx, "[WebSocket] Reconnecting, try %v after %v ms", c.reconnectAttempt, delay)
	c.reconnectAttempt++
	c.reconnectTimer = time.AfterFunc(time.Millisecond*time.Duration(delay), func() {
		c.mu.Lock()
		err := c.connectUnsafe(ctx)
		if err == nil {
			c.isReconnecting = false
			c.reconnectAttempt = 0
			c.reconnectTimer = nil
		} else {
			c.isReconnecting = false
			c.reconnectUnsafe(ctx, false)
		}
		c.mu.Unlock()
	})
}

func (c *defaultWsClient) min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (c *defaultWsClient) max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (c *defaultWsClient) getReconnectDelay() int {
	if c.reconnectAttempt <= 1 {
		return 250 + randIntn(250)
	}
	if c.reconnectAttempt <= 5 {
		return 750 + randIntn(500)
	}
	min := c.min(c.max(750, (c.reconnectAttempt-5-1)*2000), 10000)
	max := c.min(1000+(c.reconnectAttempt-5)*2000, 15000)
	return randIntn(max-min) + min
}

func (c *defaultWsClient) startHealthCheckUnsafe(ctx context.Context) {
	if c.healthCheckTicker != nil {
		c.healthCheckTicker.Stop()
	}
	ticker := time.NewTicker(c.healthCheckInterval)
	c.healthCheckTicker = ticker
	go func() {
		for range ticker.C {
			_ = c.sendMessage(ctx, &model.WebSocketMessage{
				Content: &model.WebSocketMessage_Ping_{
					Ping: &model.WebSocketMessage_Ping{
						Timestamp: getCurrentTimestamp(),
					},
				},
			})
		}
	}()
}

func (c *defaultWsClient) startReconnectCheckUnsafe(ctx context.Context) {
	if c.reconnectCheckTicker != nil {
		c.reconnectCheckTicker.Stop()
	}
	ticker := time.NewTicker(c.reconnectCheckInterval)
	c.reconnectCheckTicker = ticker
	go func() {
		for range ticker.C {
			c.mu.Lock()
			if c.lastMessageAt != 0 {
				duration := getSystemTimestamp() - c.lastMessageAt
				if duration > (c.aliveTimeout.Nanoseconds() / 1000000) {
					c.reconnectUnsafe(ctx, false)
				}
			}
			c.mu.Unlock()
		}
	}()
}

func (c *defaultWsClient) clearTimerUnsafe() {
	if c.healthCheckTicker != nil {
		c.healthCheckTicker.Stop()
		c.healthCheckTicker = nil
	}
	if c.reconnectCheckTicker != nil {
		c.reconnectCheckTicker.Stop()
		c.reconnectCheckTicker = nil
	}
}

func (c *defaultWsClient) sendMessage(ctx context.Context, message *model.WebSocketMessage) error {
	c.ensurePing(ctx)

	body, _ := proto.Marshal(message)
	secureMessage, err := encryptMessage(c.secret, body)
	if err != nil {
		return err
	}
	data, _ := proto.Marshal(secureMessage)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isConnecting || c.isReconnecting || c.socket == nil {
		return errors.New("empty ws client")
	}
	_ = c.socket.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	if err := c.socket.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return err
	}

	return nil
}

func (c *defaultWsClient) handleMessage(ctx context.Context, socket *websocket.Conn) {
	process := func(data []byte) error {
		var secureMessage model.SecureMessage
		err := proto.Unmarshal(data, &secureMessage)
		if err != nil {
			return err
		}

		body, err := decryptMessage(c.secret, &secureMessage)
		if err != nil {
			return err
		}

		var message model.WebSocketMessage
		err = proto.Unmarshal(body, &message)
		if err != nil {
			return err
		}

		switch content := message.Content.(type) {
		case *model.WebSocketMessage_Pong_:
			{
				setServerTimeBase(content.Pong.Timestamp)
			}
		case *model.WebSocketMessage_Event_:
			{
				header := content.Event.EventHeader
				v, ok := c.eventHandlerMap.Load(header.EventType)
				if !ok {
					_, loaded := c.eventUnsupportType.LoadOrStore(header.EventType, true)
					if !loaded {
						c.config.Logger.Warnf(ctx, "Unhandled event type %s", header.EventType)
					}
				} else {
					set := v.(*defaultWsEventHandlerSet)
					set.mu.RLock()
					handlers := set.handlers
					set.mu.RUnlock()
					for _, handler := range handlers {
						err = handler(ctx, &EventHeader{
							EventId:        header.EventId,
							EventType:      header.EventType,
							EventCreatedAt: header.EventCreatedAt,
						}, content.Event.EventBody)
						if err != nil {
							return err
						}
					}
				}
			}
		case *model.WebSocketMessage_HttpResponse:
			{
				resp := content.HttpResponse
				c.reqCallbacksMu.Lock()
				cb := c.reqCallbacks[resp.ReqId]
				delete(c.reqCallbacks, resp.ReqId)
				c.reqCallbacksMu.Unlock()
				if cb != nil {
					cb.resp <- resp
				}
			}
		}

		return nil
	}

	err := func() error {
		for {
			_, data, err := socket.ReadMessage()
			if err != nil {
				return err
			}

			c.mu.Lock()
			if c.socket != socket {
				c.mu.Unlock()
				return errors.New("invalid socket")
			}
			c.lastMessageAt = getSystemTimestamp()
			c.mu.Unlock()

			go func() {
				if err := process(data); err != nil {
					c.config.Logger.Errorf(ctx, "[WebSocket] handleMessage error: %v", err)
				}
			}()
		}
	}()
	if err != nil {
		_ = socket.Close()
	}
}

func (c *defaultWsClient) newReqId() string {
	c.reqCount++
	return fmt.Sprint(c.reqCount)
}
