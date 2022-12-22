// router of vpnHelper
// Powered By Luckykeeper <luckykeeper@luckykeeper.site | https://luckykeeper.site>
package router

import (
	"net/http"
	handler "vpnHelper/handler"

	"github.com/gin-gonic/gin"
)

func GinRouter(router *gin.Engine) {
	// LOGO
	router.StaticFile("/favicon.ico", "./static/favicon.ico")
	router.StaticFile("/static/cocoa", "./static/cocoa")
	router.StaticFile("/static/ba", "./static/ba")
	router.StaticFile("/static/sakura.png", "./static/sakura.png")
	router.StaticFile("/static/t9kX2ZsbR71DFh0IscwBnjtRgJVuakEywhN",
		"./static/t9kX2ZsbR71DFh0IscwBnjtRgJVuakEywhN")
	router.LoadHTMLGlob("templates/*")

	// 根路径
	router.Any("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{})
	})

	// api - V1 - 路由组
	v1 := router.Group("/v1")

	//  api - V1 - 根路径
	v1.Any("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "v1.tmpl", gin.H{})
	})

	// api - V1 - 生成用户
	v1.POST("/generateUserProfile", handler.GenerateUserProfile)

	// api - V1 - 获取用户连接信息(节点信息合并返回)
	v1.POST("/getConnInfoByUser", handler.GetConnInfoByUser)

	// api - V1 - 删除用户
	v1.POST("/delUserProfile", handler.DelUserProfile)
}
