package router

import (
	"regexp"
	"strings"
)

type RouteType int

const (
	RouteTypeStatic   RouteType = iota // 静态路由，如 /api/users
	RouteTypeParam                     // 参数路由(正则匹配)，如 /api/users/:id(\d+)
	RouteTypeWildcard                  // 通配符路由，如 /static/*
)

type RoutePattern struct {
	Original string
	Regex    *regexp.Regexp
	Type     RouteType
	Keys     []string // 参数名
}

func ParseRoute(pattern string) *RoutePattern {
	rp := &RoutePattern{
		Original: pattern,
		Keys:     make([]string, 0),
	}

	switch {
	case strings.Contains(pattern, "/*"):
		rp.Type = RouteTypeWildcard
		base := strings.TrimSuffix(pattern, "/*")
		if base == "" {
			base = "/"
		}
		rp.Regex = regexp.MustCompile("^" + regexp.QuoteMeta(base) + "/.*$")

	case strings.Contains(pattern, "/:"):
		rp.Type = RouteTypeParam
		regexStr := "^"
		keyRegex := regexp.MustCompile(`/:(\w+)(?:\(([^)]+)\))?`)
		matches := keyRegex.FindAllStringSubmatch(pattern, -1)
		lastPos := 0

		for _, match := range matches {
			staticPart := pattern[lastPos:strings.Index(pattern, match[0])]
			regexStr += regexp.QuoteMeta(staticPart)

			paramName := match[1]
			paramPattern := "[^/]+"
			if match[2] != "" {
				paramPattern = match[2]
			}

			regexStr += "(?P<" + paramName + ">" + paramPattern + ")"
			rp.Keys = append(rp.Keys, paramName)
			lastPos = strings.Index(pattern, match[0]) + len(match[0])
		}

		if lastPos < len(pattern) {
			regexStr += regexp.QuoteMeta(pattern[lastPos:])
		}
		regexStr += "$"
		rp.Regex = regexp.MustCompile(regexStr)

	default:
		rp.Type = RouteTypeStatic
		rp.Regex = regexp.MustCompile("^" + regexp.QuoteMeta(pattern) + "$")
	}

	return rp
}

type RouterGroup struct {
	staticRoutes   []*RoutePattern
	paramRoutes    []*RoutePattern
	wildcardRoutes []*RoutePattern
}

func NewRouterGroup() *RouterGroup {
	return &RouterGroup{
		staticRoutes:   make([]*RoutePattern, 0),
		paramRoutes:    make([]*RoutePattern, 0),
		wildcardRoutes: make([]*RoutePattern, 0),
	}
}

func (r *RouterGroup) AddRoute(path string) {
	route := ParseRoute(path)

	switch route.Type {
	case RouteTypeStatic:
		r.staticRoutes = append(r.staticRoutes, route)

	case RouteTypeParam:
		r.paramRoutes = append(r.paramRoutes, route)

	case RouteTypeWildcard:
		r.wildcardRoutes = append(r.wildcardRoutes, route)
	}
}

//func (r *RouterGroup) Match(method, path string) (fasthttp.RequestHandler, map[string]string) {
//	// 1. 检查静态路由
//	key := method + " " + path
//	if handler, ok := r.staticRoutes[key]; ok {
//		return handler, nil
//	}
//
//	// 2. 检查参数路由
//	for _, route := range r.paramRoutes {
//		if matches := route.Regex.FindStringSubmatch(path); matches != nil {
//			params := make(map[string]string)
//			for i, name := range route.Regex.SubexpNames() {
//				if i > 0 && i <= len(matches) && name != "" {
//					params[name] = matches[i]
//				}
//			}
//			return route.ProtocolName, params
//		}
//	}
//
//	// 3. 检查通配符路由
//	for _, route := range r.wildcardRoutes {
//		if route.Regex.MatchString(path) {
//			params := map[string]string{
//				"*": strings.TrimPrefix(path, strings.TrimSuffix(route.Original, "/*")),
//			}
//			return route.ProtocolName, params
//		}
//	}
//
//	return nil, nil
//}
