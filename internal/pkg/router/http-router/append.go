package http_router

import (
	"go-faster-gateway/internal/pkg/router"
	"go-faster-gateway/pkg/checker"
	"net/url"
	"sort"
	"strings"

	"github.com/valyala/fasthttp"
)

type RuleType = string

const (
	HttpHeader RuleType = "header"
	HttpQuery  RuleType = "query"
	HttpCookie RuleType = "cookie"
)

func Parse(rules []router.AppendRule) router.MatcherChecker {
	if len(rules) == 0 {
		return &router.EmptyChecker{}
	}
	rls := make(router.RuleCheckers, 0, len(rules))

	for _, r := range rules {
		ck, _ := checker.Parse(r.Pattern)

		switch strings.ToLower(r.Type) {
		case HttpHeader:
			rls = append(rls, &HeaderChecker{
				name:    r.Name,
				Checker: ck,
			})
		case HttpQuery:
			rls = append(rls, &QueryChecker{
				name:    r.Name,
				Checker: ck,
			})
		case HttpCookie:
			rls = append(rls, &CookieChecker{
				name:    r.Name,
				Checker: ck,
			})
		}
	}
	sort.Sort(rls)
	return rls
}

type HeaderChecker struct {
	name string
	checker.Checker
}

func (h *HeaderChecker) Weight() int {
	return int(checker.CheckTypeAll-h.Checker.CheckType()) * len(h.Checker.Value())
}

func (h *HeaderChecker) MatchCheck(req interface{}) bool {
	request, ok := req.(fasthttp.RequestCtx)
	if !ok {
		return false
	}
	v := string(request.Request.Header.Peek(h.name))
	has := len(v) > 0
	return h.Checker.Check(v, has)
}

type CookieChecker struct {
	name string
	checker.Checker
}

func (c *CookieChecker) Weight() int {
	return int(checker.CheckTypeAll-c.Checker.CheckType()) * len(c.Checker.Value())
}

func (c *CookieChecker) MatchCheck(req interface{}) bool {
	request, ok := req.(fasthttp.RequestCtx)
	if !ok {
		return false
	}
	v := string(request.Request.Header.Cookie(c.name))
	has := len(v) > 0
	return c.Checker.Check(v, has)
}

type QueryChecker struct {
	name string
	checker.Checker
}

func (q *QueryChecker) Weight() int {
	return int(checker.CheckTypeAll-q.Checker.CheckType()) * len(q.Checker.Value())
}

func (q *QueryChecker) MatchCheck(req interface{}) bool {
	request, ok := req.(fasthttp.RequestCtx)
	if !ok {
		return false
	}
	qs := request.URI().QueryString()
	values, err := url.ParseQuery(string(qs))
	if err != nil {
		return false
	}
	v := values.Get(q.Checker.Value())
	has := len(v) > 0
	return q.Checker.Check(v, has)
}
