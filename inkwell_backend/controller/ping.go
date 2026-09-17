package controller

import "github.com/gin-gonic/gin"

// PingHandler 健康检查, 用于探活/负载均衡检查
//
// @Summary  健康检查
// @Description  探活接口, 游客可访问
// @Tags  系统
// @Produce  json
// @Success  200  {object}  controller.ResponseData{data=string}  "服务正常, data 固定为 pong"
// @Router  /ping [get]
func PingHandler(c *gin.Context) {
	ResponseSuccess(c, "pong")
}
