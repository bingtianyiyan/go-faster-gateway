package protocols

import (
	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/log"
	"strings"
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
	conn     *websocket.Conn
	send     chan []byte
	userID   string
	LastPing time.Time
}

//// 使用 haxmap 实现客户端管理
//var (
//	clients = haxmap.New[string, *Client]() // key: UserID, value: *Client
//	mutex   sync.RWMutex
//)

type WSHandler struct {
	clients map[*websocket.Conn]*Client
}

func NewWSHandler() *WSHandler {
	return &WSHandler{
		clients: make(map[*websocket.Conn]*Client),
	}
}

func (h *WSHandler) Handle(ctx *fasthttp.RequestCtx, routerInfo *dynamic.ServiceRoute) {
	// 中间件在WebSocket升级前执行
	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		defer conn.Close()

		// 从请求中获取用户ID（示例用Header，实际可改用JWT等）
		userID := string(ctx.Request.Header.Peek("X-User-ID"))
		if userID == "" {
			//mock
			userID = "123"
			//ctx.Error("Unauthorized", fasthttp.StatusUnauthorized)
			//return
		}
		// 创建客户端对象
		client := &Client{
			conn:     conn,
			userID:   userID,
			LastPing: time.Now(),
		}

		// 注册到全局缓存
		h.clients[conn] = client
		defer delete(h.clients, conn) // 断开时自动注销

		// 消息处理循环
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
					log.Log.Infof("Client %s disconnected abnormally: %v", client.userID, err)
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
		client.LastPing = time.Now()
		return
	}

	// 业务逻辑（示例：广播消息）
	log.Log.Infof("Received from %s: %s", client.userID, msg)
	h.broadcastMessage(msg)
}

// 广播消息给所有客户端
func (h *WSHandler) broadcastMessage(msg []byte) {
	//clients.Range(func(key, value interface{}) bool {
	//	client := value.(*Client)
	//	if err := client.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
	//		log.Log.Infof("Failed to send to %s: %v", client.userID, err)
	//	}
	//	return true
	//})

	for k, _ := range h.clients {
		if err := k.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Log.Infof("Failed to send to %s: %v", k.RemoteAddr(), err)
		}
	}
}

// 心跳检测（自动清理断连客户端）
func (h *WSHandler) checkHeartbeat() {
	ticker := time.NewTicker(300 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for k, v := range h.clients {
			if time.Since(v.LastPing) > 60*time.Second {
				log.Log.Infof("Client %s heartbeat timeout", v.conn.RemoteAddr())
				//client.conn.Close(websocket.CloseGoingAway, "No heartbeat")
				k.Close()
				delete(h.clients, k)
			}
		}

	}
}
