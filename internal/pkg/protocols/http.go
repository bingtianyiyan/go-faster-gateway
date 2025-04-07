package protocols

import (
	"errors"
	"github.com/valyala/fasthttp"
	"go-faster-gateway/internal/pkg/balancer"
	"go-faster-gateway/internal/pkg/ecode"
	"go-faster-gateway/pkg/config/dynamic"
	"go-faster-gateway/pkg/log"
	"io"
	"strings"
	"time"
)

var _ ProtocolHandler = (*HTTPHandler)(nil)

type HTTPHandler struct {
	upstreamManager *balancer.UpstreamManager
}

func NewHTTPHandler(upstreamManager *balancer.UpstreamManager) *HTTPHandler {
	return &HTTPHandler{
		upstreamManager: upstreamManager,
	}
}

func (h *HTTPHandler) Handle(ctx *fasthttp.RequestCtx, serviceRoute *dynamic.ServiceRoute, routeInfo dynamic.Router) {
	if strings.ToLower(string(ctx.Request.Header.Peek("Upgrade"))) == "websocket" {
		return // WebSocket请求交给WebSocket处理器
	}

	if ctx.Err() != nil {
		return
	}

	// 处理完成，转发至后端服务
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	// 复制客户端请求的数据
	ctx.Request.CopyTo(req)
	// 2. 复制原始请求头和Body
	ctx.Request.Header.CopyTo(&req.Header)
	// 2. 流式转发原始请求体
	if ctx.Request.IsBodyStream() {
		req.SetBodyStream(ctx.Request.BodyStream(), ctx.Request.Header.ContentLength())
	} else {
		req.SetBody(ctx.Request.Body())
	}

	// 获取负载均衡地址
	upstreamServer, err := h.upstreamManager.GetLBUpstream(serviceRoute.RouteName, serviceRoute)
	if err != nil {
		ctx.Error(err.Error(), ecode.InternalServerErrorErr.Code)
		return
	}
	proxy := fasthttp.HostClient{
		Addr:  upstreamServer,
		IsTLS: ctx.IsTLS(),
	}
	var proxyPath string
	//TODO match route
	proxyPath = string(ctx.Path())

	req.SetRequestURI("http://" + upstreamServer + proxyPath)
	// 创建一个新的响应
	resp := fasthttp.AcquireResponse()
	//defer fasthttp.ReleaseResponse(resp)

	// 向目标后端服务器发送请求
	err = proxy.DoTimeout(req, resp, time.Second*5)
	if err != nil {
		log.Log.WithError(err).Error("fasthttp.doTimeout()")
		if errors.Is(err, fasthttp.ErrTimeout) {
			ctx.Error(ecode.BackendTimeoutErr.Data(), ecode.BackendTimeoutErr.HttpCode)
		} else {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		}
		return
	}

	defer func() {
		if resp.IsBodyStream() {
			if rs, ok := resp.BodyStream().(io.Closer); ok {
				rs.Close() // 确保流关闭
			}
		}
	}()

	// 将目标服务器的响应返回给客户端
	// 将目标服务器的响应头部和主体复制到当前请求对象中
	resp.Header.CopyTo(&ctx.Response.Header)
	// 流式响应处理
	if resp.IsBodyStream() {
		ctx.SetBodyStream(&safeReadCloser{
			ReadCloser: resp.BodyStream().(io.ReadCloser),
			onClose: func() {
				fasthttp.ReleaseResponse(resp) // 延迟释放
			},
		}, resp.Header.ContentLength())
	} else {
		ctx.Response.SetBody(resp.Body())
		fasthttp.ReleaseResponse(resp) // 立即释放
	}
	//if ctx.Request.IsBodyStream() {
	//	ctx.SetBodyStream(resp.BodyStream(), resp.Header.ContentLength())
	//} else {
	//	ctx.Response.SetBody(resp.Body())
	//}
}

func (h *HTTPHandler) Supports(ctx *fasthttp.RequestCtx) bool {
	return !(strings.ToLower(string(ctx.Request.Header.Peek("Upgrade"))) == "websocket") &&
		(ctx.IsGet() || ctx.IsPost() || ctx.IsDelete() || ctx.IsPut() || ctx.IsOptions() || ctx.IsHead() || ctx.IsTrace())

}

type safeReadCloser struct {
	io.ReadCloser
	onClose func()
}

func (rc *safeReadCloser) Close() error {
	err := rc.ReadCloser.Close()
	rc.onClose()
	return err
}
