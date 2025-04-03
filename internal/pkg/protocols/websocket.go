package protocols

import (
	"fmt"
	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
	"go-faster-gateway/internal/pkg/balancer"
	"go-faster-gateway/internal/pkg/ecode"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/helper/env"
	"go-faster-gateway/pkg/log"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 1024 * 8 * 5 //5g
)

var upgrader = websocket.FastHTTPUpgrader{
	HandshakeTimeout: 5 * time.Second, // 快速握手
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		return true
		//origin := string(ctx.Request.Header.Peek("Origin"))
		//allowedOrigins := []string{"https://yourdomain.com"}
		//for _, o := range allowedOrigins {
		//	if origin == o {
		//		return true
		//	}
		//}
	}, // 允许所
}

type Client struct {
	hub        *MessageCenter
	conn       *websocket.Conn
	send       chan []byte
	lastPing   time.Time
	backendURL string // 对应的后端服务地址
}

type ClientMsg struct {
	msg    []byte
	client *Client
}

type WSHandler struct {
	upstreamManager *balancer.UpstreamManager
	msgCenter       *MessageCenter
	counter         int
	mu              sync.Mutex // 声明互斥锁
}

func NewWSHandler(upstreamManager *balancer.UpstreamManager) *WSHandler {
	return &WSHandler{
		upstreamManager: upstreamManager,
		msgCenter:       newMessageCenter(),
	}
}

func (h *WSHandler) Handle(ctx *fasthttp.RequestCtx, serviceRoute *dynamic.ServiceRoute, routeInfo dynamic.Router) {
	h.mu.Lock()         // 加锁
	defer h.mu.Unlock() // 确保解锁(即使发生panic)
	if h.counter == 0 {
		go h.msgCenter.run()
	}

	// 中间件在WebSocket升级前执行
	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		defer conn.Close()

		// 获取负载均衡地址
		upstreamServer, err := h.upstreamManager.GetLBUpstream(serviceRoute.RouteName, serviceRoute)
		if err != nil {
			ctx.Error(err.Error(), ecode.InternalServerErrorErr.Code)
			return
		}
		//这边先默认只配置一个websocket的/ws地址
		backendURL := fmt.Sprintf("%s%s", upstreamServer, serviceRoute.RouteGroup+routeInfo.Prefix+routeInfo.Path)
		client := &Client{hub: h.msgCenter, conn: conn, send: make(chan []byte, 256), backendURL: backendURL, lastPing: time.Now()}
		client.hub.register <- client
		//收到消息处理
		go client.writePump()
		//读取websocket消息转发到消息中心
		client.readPump()
	})

	if err != nil {
		ctx.Error("WebSocket upgrade failed", fasthttp.StatusBadRequest)
	}

	// 启动心跳检测协程
	if h.counter == 0 {
		go h.checkHeartbeat()
	}
	h.counter++
}

func (h *WSHandler) Supports(ctx *fasthttp.RequestCtx) bool {
	if strings.ToLower(string(ctx.Request.Header.Peek("Upgrade"))) == "websocket" {
		return true // WebSocket请求交给WebSocket处理器
	}
	return false
}

// 心跳检测（自动清理断连客户端）
func (h *WSHandler) checkHeartbeat() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		for client, _ := range h.msgCenter.clients {
			if now.Sub(client.lastPing) > 300*time.Second {
				log.Log.Infof("Client %s heartbeat timeout", client.conn.RemoteAddr())
				h.msgCenter.unregister <- client
			}
		}

	}
}

// readPump pumps messages from the websocket connection to the msgCenter
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		msgType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Log.WithError(err).Error("c.conn.ReadMessage fail")
			}
			break
		}

		go c.processMessage(msgType, message)
	}
}

// writePump pumps messages from the msgCenter to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			//转发到后端服务处理
			go c.ForwardToBackend(message)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// 异步消息处理
func (s *Client) processMessage(msgType int, msg []byte) {
	// 更新心跳时间
	if msgType == websocket.PingMessage {
		s.lastPing = time.Now()
		s.hub.clients[s] = true
		return
	}

	// 业务逻辑（示例：广播消息）
	log.Log.Debugf("Received from %s %s", s.conn.RemoteAddr(), msg)
	msg2 := &ClientMsg{
		msg:    msg,
		client: s,
	}
	s.hub.singlebroadcast <- msg2
}

// 转发消息到后端服务
func (s *Client) ForwardToBackend(message []byte) {
	//环境变量判断
	envDefaultName := env.ModeDebug.String()
	envName := os.Getenv("EnvName")
	//测试启动
	if envName == envDefaultName {
		msg := "server replay" + string(message)
		if err := s.conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Log.WithError(err).Error("Client write error:")
		}
		log.Log.Debug("replay msg to client")
		return
	} else {
		// 1. 建立到后端服务的WebSocket连接
		backendConn, _, err := websocket.DefaultDialer.Dial(s.backendURL, nil)
		if err != nil {
			log.Log.Infof("Failed to connect to backend: %v", err)
			return
		}
		defer backendConn.Close()

		// 2. 转发消息
		if err := backendConn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Log.WithError(err).Error("Backend write error")
			return
		}

		// 3. 接收后端响应（可选）
		_, resp, err := backendConn.ReadMessage()
		if err != nil {
			log.Log.WithError(err).Error("Backend read error")
			return
		}

		// 4. 将响应返回给客户端
		if err := s.conn.WriteMessage(websocket.TextMessage, resp); err != nil {
			log.Log.WithError(err).Error("Client write error:")
		}
	}
}

// 消息中心
type MessageCenter struct {
	clients map[*Client]bool

	//单播消息
	singlebroadcast chan *ClientMsg

	// 注册客户端
	register chan *Client

	// 注销客户端
	unregister chan *Client
}

func newMessageCenter() *MessageCenter {
	return &MessageCenter{
		singlebroadcast: make(chan *ClientMsg),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		clients:         make(map[*Client]bool),
	}
}

func (h *MessageCenter) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				log.Log.Infof("unregister client:%s", client.conn.RemoteAddr())
				delete(h.clients, client)
				close(client.send)
			}
		case singleMsg := <-h.singlebroadcast: // 消息单播
			singleMsg.client.send <- singleMsg.msg //转发到后端

		}
	}
}
