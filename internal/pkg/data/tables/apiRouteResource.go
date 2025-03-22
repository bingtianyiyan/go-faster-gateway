package tables

import (
	"database/sql/driver"
	"encoding/json"
)

type ApiRouteResource struct {
	Id          uint       `gorm:"AUTO_INCREMENT"`
	ApiPath     string     `json:"api_path"`     // 网关请求路径
	Method      string     `json:"method"`       // get/post/put/...
	ServiceName string     `json:"service_name"` // 上游服务名，用于寻后端的具体服务
	ProxyPath   string     `json:"proxy_path"`   // 路由上游路径，如果需要转发至该路径
	Middleware  Middleware `json:"middleware"`   // 一组中间件集合，路由的整体生命周期里，需要经过此批中间件的一层层处理
}
type Middleware []string

func (t *Middleware) Scan(value interface{}) error {
	bytesValue, _ := value.([]byte)
	return json.Unmarshal(bytesValue, t)
}

func (t *Middleware) Value() (driver.Value, error) {
	return json.Marshal(t)
}
func (ApiRouteResource) TableName() string {
	return "api"
}
