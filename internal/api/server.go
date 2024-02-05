package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var Server HttpServer

//http服务大类
type HttpServer struct {
	RoutePrefix string
	server      *http.Server
	tlsCert     string
	tlsKey      string
	bind        string
	handler     *gin.Engine
}

//响应web界面端
type ResultMessage struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

