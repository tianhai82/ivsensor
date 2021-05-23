package main

import (
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/tianhai82/ivsensor/doc"
	"github.com/tianhai82/ivsensor/tda"
	"github.com/tianhai82/ivsensor/webapi"
)

func main() {
	println("starting ivsensor...")
	r := gin.Default()
	r.Use(static.Serve("/", static.LocalFile("./web/dist", false)))

	webRouter := r.Group("/api/web")
	webapi.AddApi(webRouter)

	tdaRouter := r.Group("/api/tda")
	tda.AddApi(tdaRouter)
	// r.GET("/api/crawlOptions", crawler.HandleCrawlOption)
	r.GET("/doc/:filename", doc.HandleDownload)
	// r.GET("/gendoc/:date", doc.GenDoc)
	r.Run()
}
