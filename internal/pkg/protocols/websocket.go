package protocols

import (
	"fmt"
	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
	"go-faster-gateway/internal/pkg/balancer"
	"go-faster-gateway/internal/pkg/ecode"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/log"
	"strings"
	"sync"
	"time"
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
	conn       *websocket.Conn
	send       chan []byte
	lastPing   time.Time
	backendURL string // 对应的后端服务地址
}

type WSHandler struct {
	upstreamManager *balancer.UpstreamManager
	lockMap         sync.Map
	// 保护后端转发连接的锁
	backendLock sync.Mutex
}

func NewWSHandler(upstreamManager *balancer.UpstreamManager) *WSHandler {
	return &WSHandler{
		upstreamManager: upstreamManager,
	}
}

func (h *WSHandler) Handle(ctx *fasthttp.RequestCtx, serviceRoute *dynamic.ServiceRoute, routeInfo dynamic.Router) {
	// 中间件在WebSocket升级前执行
	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		defer conn.Close()

		// 创建客户端对象
		// 注册到全局缓存
		client, err := h.AddClient(ctx, conn, serviceRoute, routeInfo)
		if err != nil {
			log.Log.WithError(err).Error("addClient fail")
			return
		}
		defer h.RemoveClient(conn)

		// 消息处理循环
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
					log.Log.Infof("Client %s disconnected abnormally: %v", client.conn.RemoteAddr(), err)
				}
				break
			}

			// 异步处理消息（避免阻塞）
			go h.processMessage(client, msgType, msg)
		}
	})

	if err != nil {
		ctx.Error("WebSocket upgrade failed", fasthttp.StatusBadRequest)
	}

	// 启动心跳检测协程
	go h.checkHeartbeat()
}

func (h *WSHandler) Supports(ctx *fasthttp.RequestCtx) bool {
	if strings.ToLower(string(ctx.Request.Header.Peek("Upgrade"))) == "websocket" {
		return true // WebSocket请求交给WebSocket处理器
	}
	return false
}

// 异步消息处理
func (h *WSHandler) processMessage(client *Client, msgType int, msg []byte) {
	// 更新心跳时间
	if msgType == websocket.PingMessage {
		client.lastPing = time.Now()
		h.lockMap.LoadOrStore(client.conn, client)
		return
	}

	// 业务逻辑（示例：广播消息）
	log.Log.Debugf("Received from %s %s", client.conn.RemoteAddr(), msg)
	//h.broadcastMessage(msg)
	h.ForwardToBackend(client, msg)
}

// 广播消息给所有客户端
func (h *WSHandler) broadcastMessage(msg []byte) {
	h.lockMap.Range(func(key, value any) bool {
		client := value.(*Client)
		if err := client.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Log.Infof("Failed to send to %s: %v", client.conn.RemoteAddr(), err)
		}
		return true
	})
}

// 转发消息到后端服务
func (h *WSHandler) ForwardToBackend(client *Client, message []byte) {
	h.backendLock.Lock()
	defer h.backendLock.Unlock()

	// 1. 建立到后端服务的WebSocket连接
	backendConn, _, err := websocket.DefaultDialer.Dial(client.backendURL, nil)
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
	if err := client.conn.WriteMessage(websocket.TextMessage, resp); err != nil {
		log.Log.WithError(err).Error("Client write error:")
	}
}

// 心跳检测（自动清理断连客户端）
func (h *WSHandler) checkHeartbeat() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		h.lockMap.Range(func(key, value any) bool {
			client := value.(*Client)
			if now.Sub(client.lastPing) > 300*time.Second {
				log.Log.Infof("Client %s heartbeat timeout", client.conn.RemoteAddr())
				client.conn.Close() // 关闭失效连接
				h.RemoveClient(client.conn)
			}
			return true
		})

	}
}

// 添加客户端
func (h *WSHandler) AddClient(ctx *fasthttp.RequestCtx, conn *websocket.Conn, serviceRoute *dynamic.ServiceRoute, routeInfo dynamic.Router) (*Client, error) {
	// 获取负载均衡地址
	upstreamServer, err := h.upstreamManager.GetLBUpstream(serviceRoute.RouteName, serviceRoute)
	if err != nil {
		ctx.Error(err.Error(), ecode.InternalServerErrorErr.Code)
		return nil, err
	}
	//这边先默认只配置一个websocket的/ws地址
	backendURL := fmt.Sprintf("%s%s", upstreamServer, serviceRoute.RouteGroup+routeInfo.Prefix+routeInfo.Path)
	client := &Client{
		conn:       conn,
		lastPing:   time.Now(),
		backendURL: backendURL,
	}
	h.lockMap.Store(conn, client)
	return client, err
}

// 删除客户端
func (h *WSHandler) RemoveClient(conn *websocket.Conn) {
	h.lockMap.Delete(conn)
}

// 获取客户端
func (h *WSHandler) GetClient(conn *websocket.Conn) (*Client, bool) {
	data, ok := h.lockMap.Load(conn)
	if ok {
		return data.(*Client), ok
	}
	return nil, ok
}
