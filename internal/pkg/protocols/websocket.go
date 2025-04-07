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
	maxMessageSize = 10 << 20 // 10MB
)

var (
	//客户端->后端
	clientToBack = 1
	backToClient = 2
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

type WSHandler struct {
	upstreamManager *balancer.UpstreamManager
	sessionCenter   *SessionCenter
	once            sync.Once
	once_heart      sync.Once
}

func NewWSHandler(upstreamManager *balancer.UpstreamManager) *WSHandler {
	return &WSHandler{
		upstreamManager: upstreamManager,
		sessionCenter:   newSessionCenter(),
	}
}

// 直接使用fasthttp websocket
func (h *WSHandler) Handle(ctx *fasthttp.RequestCtx, serviceRoute *dynamic.ServiceRoute, routeInfo dynamic.Router) {
	h.once.Do(func() {
		go h.sessionCenter.run()
	})

	// 中间件在WebSocket升级前执行
	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		var backendConn *websocket.Conn
		defer func() {
			if conn != nil {
				defer conn.Close()
			}
			if backendConn != nil {
				defer backendConn.Close()
			}
		}()

		// 获取负载均衡地址
		upstreamServer, err := h.upstreamManager.GetLBUpstream(serviceRoute.RouteName, serviceRoute)
		if err != nil {
			ctx.Error(err.Error(), ecode.InternalServerErrorErr.Code)
			return
		}
		//这边先默认只配置一个websocket的/ws地址
		path := routeInfo.ProxyPath
		if len(path) == 0 {
			path = routeInfo.Path
		}

		backendURL := fmt.Sprintf("%s%s", upstreamServer, serviceRoute.RouteGroup+routeInfo.Prefix+path)

		//环境变量判断
		envDefaultName := env.ModeDebug.String()
		envName := os.Getenv("EnvName")
		if envDefaultName != envName {
			backendConn, _, err = websocket.DefaultDialer.Dial(backendURL, nil)
			if err != nil {
				log.Log.Infof("Failed to connect to backend: %v", err)
				return
			}
		} else {
			backendConn = conn
		}

		sessionClient := &SessionClient{
			sessionHub: h.sessionCenter,
			conn:       conn,
			backConn:   backendConn,
			lastPing:   time.Now(),
		}
		h.sessionCenter.register <- sessionClient
		go sessionClient.writePump()
		err = h.proxyWS(sessionClient)
		if err != nil {
			ctx.Error(err.Error(), ecode.InternalServerErrorErr.Code)
			return
		}
	})

	if err != nil {
		ctx.Error("WebSocket upgrade failed", fasthttp.StatusBadRequest)
	}
	h.once_heart.Do(func() {
		go h.checkHeartbeat()
	})
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
		for _, client := range h.sessionCenter.clients {
			if now.Sub(client.lastPing) > 300*time.Second {
				log.Log.Infof("Client %s heartbeat timeout", client.conn.RemoteAddr())
				h.sessionCenter.unregister <- client
			}
		}

	}
}

var bufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 32*1024) // 32KB缓冲
	},
}

func (h *WSHandler) proxyWS(sessionClient *SessionClient) error {
	errChan := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	// 客户端->后端
	go sessionClient.pipeMessages(errChan, &wg, clientToBack)
	// 后端->客户端
	go sessionClient.pipeMessages(errChan, &wg, backToClient)

	wg.Wait()
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

func (c *SessionClient) pipeMessages(errChan chan<- error, wg *sync.WaitGroup, sourceType int) {
	defer func() {
		if sourceType == clientToBack {
			c.sessionHub.unregister <- c
		}
		if c.conn != c.backConn {
			c.backConn.Close()
		}
		c.conn.Close()
	}()
	defer wg.Done()
	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)
	var src, dest *websocket.Conn
	if sourceType == clientToBack {
		src = c.conn
		dest = c.backConn
	} else {
		src = c.backConn
		dest = c.conn
	}
	src.SetReadLimit(maxMessageSize)
	src.SetReadDeadline(time.Now().Add(pongWait))
	src.SetPongHandler(func(string) error { src.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		msgType, msg, err := src.ReadMessage()
		if err != nil {
			if !isNormalClose(err) {
				errChan <- err
			}
			return
		}
		//如果是自身服务ping的消息则不需要往下再发送回去
		if msgType == websocket.PingMessage {
			if sourceType == clientToBack {
				c.sessionHub.updateClientHeart <- src
			}
			continue
		}

		// 动态扩容
		if len(msg) > cap(buf) {
			buf = make([]byte, len(msg))
		} else {
			buf = buf[:len(msg)]
		}
		copy(buf, msg)

		if err = dest.WriteMessage(msgType, buf); err != nil {
			errChan <- err
			return
		}
	}
}

func (c *SessionClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func isNormalClose(err error) bool {
	return websocket.IsCloseError(err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseNoStatusReceived)
}

// 会话中心
type SessionCenter struct {
	clients map[*websocket.Conn]*SessionClient

	// 注册客户端
	register chan *SessionClient

	// 注销客户端
	unregister chan *SessionClient

	//更新客户端心跳
	updateClientHeart chan *websocket.Conn
}

type SessionClient struct {
	sessionHub *SessionCenter
	conn       *websocket.Conn
	backConn   *websocket.Conn
	lastPing   time.Time
}

func newSessionCenter() *SessionCenter {
	return &SessionCenter{
		register:          make(chan *SessionClient),
		unregister:        make(chan *SessionClient),
		clients:           make(map[*websocket.Conn]*SessionClient),
		updateClientHeart: make(chan *websocket.Conn),
	}
}

func (h *SessionCenter) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.conn] = client
			log.Log.Infof("register count %s", len(h.clients))
		case client := <-h.unregister:
			if _, ok := h.clients[client.conn]; ok {
				log.Log.Infof("unregister client:%s", client.conn.RemoteAddr())
				delete(h.clients, client.conn)
			}
		case clientConn := <-h.updateClientHeart:
			if client, ok := h.clients[clientConn]; ok {
				log.Log.Infof("update client:%s", clientConn.RemoteAddr())
				client.lastPing = time.Now()
				h.clients[clientConn] = client
			}
		}
	}
}

////////////////////////// 以msgCenter模式处理消息

//type Client struct {
//	hub        *MessageCenter
//	conn       *websocket.Conn
//	backConn   *websocket.Conn
//	send       chan SendMsg
//	lastPing   time.Time
//}
//
//type SendMsg struct {
//	msgType int
//	msg     []byte
//}
//
//type ClientMsg struct {
//	msgType int
//	msg     []byte
//	client  *Client
//}
//
//
//// readPump pumps messages from the websocket connection to the msgCenter
//func (c *Client) readPump() {
//	defer func() {
//		c.hub.unregister <- c
//		c.conn.Close()
//	}()
//	c.conn.SetReadLimit(maxMessageSize)
//	c.conn.SetReadDeadline(time.Now().Add(pongWait))
//	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
//	for {
//		msgType, message, err := c.conn.ReadMessage()
//		if err != nil {
//			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
//				log.Log.WithError(err).Error("c.conn.ReadMessage fail")
//			}
//			break
//		}
//
//		go c.processMessage(msgType, message)
//	}
//}
//
//// writePump pumps messages from the msgCenter to the websocket connection.
//func (c *Client) writePump() {
//	ticker := time.NewTicker(pingPeriod)
//	defer func() {
//		ticker.Stop()
//		c.conn.Close()
//	}()
//	for {
//		select {
//		case message, ok := <-c.send:
//			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
//			if !ok {
//				// The hub closed the channel.
//				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
//				return
//			}
//			//转发到后端服务处理
//			go c.ForwardToBackend(message)
//
//		case <-ticker.C:
//			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
//			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
//				return
//			}
//		}
//	}
//}
//
//// 异步消息处理
//func (s *Client) processMessage(msgType int, msg []byte) {
//	// 更新心跳时间
//	if msgType == websocket.PingMessage {
//		s.lastPing = time.Now()
//		s.hub.clients[s] = true
//		return
//	}
//
//	// 业务逻辑（示例：广播消息）
//	log.Log.Debugf("Received from %s %s", s.conn.RemoteAddr(), msg)
//	msg2 := &ClientMsg{
//		msgType: msgType,
//		msg:     msg,
//		client:  s,
//	}
//	s.hub.singlebroadcast <- msg2
//}
//
//// 转发消息到后端服务
//func (s *Client) ForwardToBackend(message SendMsg) {
//	//环境变量判断
//	envDefaultName := env.ModeDebug.String()
//	envName := os.Getenv("EnvName")
//	//测试启动
//	if envName == envDefaultName {
//		msg := "server replay" + string(message.msg)
//		if err := s.conn.WriteMessage(message.msgType, []byte(msg)); err != nil {
//			log.Log.WithError(err).Error("Client write error:")
//		}
//		log.Log.Debug("replay msg to client")
//		return
//	} else {
//		//转发消息到后端
//		if err := s.backConn.WriteMessage(message.msgType, message.msg); err != nil {
//			log.Log.WithError(err).Error("Backend write error")
//			return
//		}
//		// 3. 接收后端响应（可选）
//		_, resp, err := s.backConn.ReadMessage()
//		if err != nil {
//			log.Log.WithError(err).Error("Backend read error")
//			return
//		}
//
//		// 4. 将响应返回给客户端
//		if err := s.conn.WriteMessage(message.msgType, resp); err != nil {
//			log.Log.WithError(err).Error("Client write error:")
//		}
//	}
//}
//
//// 消息中心
//type MessageCenter struct {
//	clients map[*Client]bool
//
//	//单播消息
//	singlebroadcast chan *ClientMsg
//
//	// 注册客户端
//	register chan *Client
//
//	// 注销客户端
//	unregister chan *Client
//}
//
//func newMessageCenter() *MessageCenter {
//	return &MessageCenter{
//		singlebroadcast: make(chan *ClientMsg),
//		register:        make(chan *Client),
//		unregister:      make(chan *Client),
//		clients:         make(map[*Client]bool),
//	}
//}
//
//func (h *MessageCenter) run() {
//	for {
//		select {
//		case client := <-h.register:
//			h.clients[client] = true
//		case client := <-h.unregister:
//			if _, ok := h.clients[client]; ok {
//				log.Log.Infof("unregister client:%s", client.conn.RemoteAddr())
//				delete(h.clients, client)
//				close(client.send)
//			}
//		case singleMsg := <-h.singlebroadcast: // 消息单播
//			singleMsg.client.send <- SendMsg{
//				msgType: singleMsg.msgType,
//				msg:     singleMsg.msg,
//			} //转发到后端
//
//		}
//	}
//}
