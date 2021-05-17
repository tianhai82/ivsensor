package main

import (
	"github.com/gin-gonic/gin"
	"github.com/tianhai82/ivsensor/doc"
	"github.com/tianhai82/ivsensor/tda"
)

func main() {
	println("starting ivsensor...")
	r := gin.Default()
	tdaRouter := r.Group("/api/tda")
	tda.AddApi(tdaRouter)
	// r.GET("/api/crawlOptions", crawler.HandleCrawlOption)
	r.GET("/doc/:filename", doc.HandleDownload)
	// r.GET("/gendoc/:date", doc.GenDoc)
	r.Run()
}
