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


