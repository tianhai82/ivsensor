package main

import (
	"fmt"

	"github.com/tianhai82/ivsensor/crawler"
)

func main() {
	err := crawler.CrawlSymbol("BLDP")
	if err != nil {
		fmt.Println(err)
	}
}
