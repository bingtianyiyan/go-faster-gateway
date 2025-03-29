package provider

//import (
//	"context"
//	"fmt"
//	"github.com/acmestack/gorm-plus/gplus"
//	"github.com/jinzhu/copier"
//	"go-faster-gateway/internal/pkg/biz"
//	"go-faster-gateway/internal/pkg/data/tables"
//	"go-faster-gateway/pkg/config/dynamic"
//	"gorm.io/gorm"
//)
//
//type apiRouteResourceRepository struct {
//}
//
//func NewApiRouteResourceRepo() biz.ApiRouteResourceRepository {
//	return &apiRouteResourceRepository{}
//}
//
//func (api *apiRouteResourceRepository) GetApiList(ctx context.Context) ([]*biz.ApiRouteResourceDto, error) {
//	query, _ := gplus.NewQuery[tables.ApiRouteResource]()
//	list, sessionDb := gplus.SelectList[tables.ApiRouteResource](query, gplus.Session(&gorm.Session{Context: ctx}))
//	if sessionDb.Error != nil {
//		return nil, sessionDb.Error
//	}
//	for _, v := range list {
//		fmt.Println(v)
//	}
//	var ret []*biz.ApiRouteResourceDto
//	for _, v := range list {
//		temp := &biz.ApiRouteResourceDto{
//			ApiPath:     v.ApiPath,
//			Method:      v.Method,
//			RouteName: v.RouteName,
//			ProxyPath:   v.ProxyPath,
//			MiddleWare:  v.Middlewares,
//		}
//		ret = append(ret, temp)
//	}
//	return ret, nil
//}
//
//// 静态文件获取路由数据
//type apiRouteResourceFileRepository struct {
//	routeList []dynamic.RouteName
//}
//
//func NewApiRouteResourceFileRepo(routes []dynamic.RouteName) biz.ApiRouteResourceRepository {
//	return &apiRouteResourceFileRepository{
//		routeList: routes,
//	}
//}
//
//func (api *apiRouteResourceFileRepository) GetApiList(ctx context.Context) ([]*biz.ApiRouteResourceDto, error) {
//	var apiList []*biz.ApiRouteResourceDto
//	err := copier.Copy(&apiList, &api.routeList)
//	if err != nil {
//		return nil, err
//	}
//	return apiList, nil
//}
