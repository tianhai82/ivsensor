package webapi

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/gin-gonic/gin"
	"github.com/tianhai82/ivsensor/firebase"
	"github.com/tianhai82/ivsensor/tda"
)

func AddApi(router *gin.RouterGroup) {
	router.GET("/tickers", getTickers)
	router.GET("/analyse/:ticker", analyseTicker)
	router.GET("/pairscore/:ticker1/:ticker2", tickerScore)
}

func getTickers(c *gin.Context) {
	c.JSON(200, firebase.StockSymbols)
}

func tickerScore(c *gin.Context) {
	ticker1 := c.Param("ticker1")
	if strings.TrimSpace(ticker1) == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	ticker2 := c.Param("ticker2")
	if strings.TrimSpace(ticker2) == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	dateStr := time.Now().Format("2006-01-02")
	atr1 := tda.StockATR{
		Symbol: ticker1,
	}
	atr1.PopulateATR(dateStr)
	atr2 := tda.StockATR{
		Symbol: ticker2,
	}
	atr2.PopulateATR(dateStr)
	score, err := tda.PairScore(atr1.Candles, atr2.Candles)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(200, score)
}

func analyseTicker(c *gin.Context) {
	ticker := c.Param("ticker")
	if strings.TrimSpace(ticker) == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	dateStr := ""
	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		fmt.Println(err)
		dateStr = time.Now().Format("2006-01-02")
	} else {
		dateStr = time.Now().In(zone).Format("2006-01-02")
	}

	stockAtr := tda.StockATR{
		Symbol: ticker,
	}
	err = stockAtr.PopulateATR(dateStr)
	if err != nil {
		fmt.Println(ticker, "fail to populate ATR", err.Error())
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	err2 := stockAtr.RetrieveOptionPremium()
	if err2 != nil {
		fmt.Println(ticker, "fail to retrieve option premium", err2.Error())
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	c.JSON(200, stockAtr)
}
