package main

import (
	"github.com/gin-gonic/gin"
	"github.com/tianhai82/ivsensor/crawler"
)

func main() {
	println("starting ivsensor...")
	r := gin.Default()
	r.GET("/api/crawlOptions", crawler.HandleCrawlOption)
	r.Run()
}
