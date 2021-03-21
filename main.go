package main

import (
	"github.com/gin-gonic/gin"
	"github.com/tianhai82/ivsensor/crawler"
	"github.com/tianhai82/ivsensor/doc"
)

func main() {
	println("starting ivsensor...")
	r := gin.Default()
	r.GET("/api/crawlOptions", crawler.HandleCrawlOption)
	r.GET("/doc/:filename", doc.HandleDownload)
	r.GET("/gendoc/:date", doc.GenDoc)
	r.Run()
}
