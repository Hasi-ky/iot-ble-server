package api

import (
	"iot-ble-server/global/globallogger"
	"iot-ble-server/internal/config"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Start() error {
	globallogger.Log.Infoln("start http server")

	return nil
}

//配置httpapi业务
func Setup() {
	Server.bind = config.C.HttpServer.Bind
	Server.tlsCert = config.C.HttpServer.TlsCert
	Server.tlsKey = config.C.HttpServer.TlsKey
	Server.RoutePrefix = config.C.HttpServer.RoutePrefix
	ginEngine := gin.Default()
	route := ginEngine.Group(Server.RoutePrefix, []gin.HandlerFunc{
		handleHttpHeader,
		handleHttpGetMethod,
	}...)
	RegistryBleHttpApi(route)
	Server.handler = ginEngine
}

//针对请求头信息当中
func handleHttpHeader(ctx *gin.Context) {
	globallogger.Log.Debugln("<handleHttpHeader> TenantId:", ctx.Request.Header.Get("TenantId"))
	ctx.Set("user_id", ctx.Request.Header.Get("Tenantid"))
}

//针对ble的get请求进行后续处理
func handleHttpGetMethod(ctx *gin.Context) {
	if ctx.Request.Method != "GET" {
		return
	}
	pageNum, err := strconv.Atoi(ctx.Query("pageNum"))
	if err != nil {
		globallogger.Log.Debugln("<handleHttpHeader>", err)
		pageNum = 1
	}
	ctx.Set("pageNum", pageNum)
	pageSize, err := strconv.Atoi(ctx.Query("pageSize"))
	if err != nil {
		globallogger.Log.Debugln("<handleHttpHeader>", err)
		pageSize = 10
	}
	ctx.Set("pageSize", pageSize)
	globallogger.Log.Debugf("<handlerHttpHeader> pageNum:%v pageSize:%v\n", pageNum, pageSize)
}

//绑定
func RegistryBleHttpApi(r *gin.RouterGroup) {
	r.DELETE("/device/", deleteDev) //批量删除终端
	r.POST("/device/connect", connectDev)
	r.GET("/device/", searchMainService)
	r.POST("/device/scan", scanDevStart)
}
