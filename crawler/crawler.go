package crawler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/piquette/finance-go/datetime"
	"github.com/piquette/finance-go/options"
	"github.com/tianhai82/ivsensor/firebase"
)

func HandleCrawlOption(c *gin.Context) {
	for _, s := range firebase.Stocks {
		getStraddle(s.Symbol)
		time.Sleep(1 * time.Second)
	}
}

func getStraddle(ticker string) {
	println("crawling", ticker)
	straddle := options.GetStraddle(ticker)
	meta := straddle.Meta()
	for _, d := range meta.AllExpirationDates {
		dt := datetime.FromUnix(d)
		iter := options.GetStraddleP(&options.Params{
			UnderlyingSymbol: ticker,
			Expiration:       dt,
		})
		if iter.Err() != nil {
			fmt.Println(iter.Err())
			continue
		}
		t := iter.Next()
		if t {
			stra := iter.Straddle()
			if stra.Call != nil {
				fmt.Println(stra.Call.Symbol)
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}
