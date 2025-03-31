package middleware

import "sync"

var (
	//中间件集合
	MiddlewareHandlerList = []MiddlewareEventHandler{}
)

type MiddlewareManager struct {
	middlewareHandler *MiddlewareHandler
	mu                sync.Mutex
}

func NewMiddlewareManager() *MiddlewareManager {
	var m MiddlewareHandler
	m.Handler = make(map[string]MiddlewareFunc)
	return &MiddlewareManager{
		middlewareHandler: &m,
	}
}

func (r *MiddlewareManager) Add(name MiddlewareName, fc MiddlewareFunc) {
	r.middlewareHandler.Handler[name.String()] = fc
}

func (r *MiddlewareManager) Delete(name string) {
	if r.middlewareHandler != nil {
		r.mu.Lock()
		defer r.mu.Lock()
		delete(r.middlewareHandler.Handler, name)
	}
}

func (r *MiddlewareManager) Get(name string) (MiddlewareFunc, bool) {
	if r.middlewareHandler != nil {
		handler, ok := r.middlewareHandler.Handler[name]
		return handler, ok
	}
	return nil, false
}

func (r *MiddlewareManager) GetAll() map[string]MiddlewareFunc {
	list := make(map[string]MiddlewareFunc)
	if r.middlewareHandler != nil {
		for v, k := range r.middlewareHandler.Handler {
			list[v] = k
		}
	}
	return list
}
