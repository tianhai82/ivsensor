package crawler

import (
	"context"
	"fmt"
	"math"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/piquette/finance-go/chart"
	"github.com/piquette/finance-go/datetime"
	"github.com/piquette/finance-go/options"
	"github.com/tianhai82/ivsensor/firebase"
	"github.com/tianhai82/ivsensor/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleCrawlOption(c *gin.Context) {
	dateStr := time.Now().Format("2006-01-02")
	task, err := firebase.FirestoreClient.Collection("task").Doc(dateStr).Get(context.Background())
	dayTask := model.DayTask{
		ID:              dateStr,
		SymbolsStatuses: nil,
	}
	if err != nil {
		if status.Code(err) == codes.NotFound {
			fmt.Println("day task not found. creating")
			statuses := map[string]bool{}
			for _, s := range firebase.Stocks {
				statuses[s.Symbol] = false
			}
			dayTask.SymbolsStatuses = statuses
			firebase.FirestoreClient.Collection("task").Doc(dateStr).Set(context.Background(), dayTask)
		} else {
			fmt.Println("error retrieving day crawl task", err)
			return
		}
	}
	if dayTask.SymbolsStatuses == nil {
		fmt.Println("dayTask symbolsStatuses is nil.. retrieve from docSnapshot")
		err = task.DataTo(&dayTask)
		if err != nil {
			fmt.Println("unable to parse dayTask", err)
			return
		}
	}

	i := 0
	for symbol, status := range dayTask.SymbolsStatuses {
		if !status {
			fmt.Println("processing", symbol)
			crawlSymbol(symbol)
			time.Sleep(1 * time.Second)
			firebase.FirestoreClient.Collection("task").Doc(dateStr).Update(context.Background(),
				[]firestore.Update{
					{
						FieldPath: firestore.FieldPath{"symbolsStatuses", symbol},
						Value:     true,
					},
				},
			)
			i++
			if i > 1 {
				break
			}
		}
	}
	fmt.Println("done")
}

func crawlSymbol(symbol string) error {
	println("crawling", symbol)

	// q, err := quote.Get(symbol)
	// if err != nil {
	// 	fmt.Println("fail to get quote for", symbol, err)
	// }
	// latestPrice := q.RegularMarketPrice

	straddle := options.GetStraddle(symbol)
	meta := straddle.Meta()
	now := time.Now()
	for _, d := range meta.AllExpirationDates {
		dt := datetime.FromUnix(d)
		days := dt.Time().Sub(now).Hours() / 24
		dte := int(math.Ceil(days))
		if dte > 60 || dte <= 0 {
			continue
		}

		iter := options.GetStraddleP(&options.Params{
			UnderlyingSymbol: symbol,
			Expiration:       dt,
		})
		if iter.Err() != nil {
			fmt.Println(iter.Err())
			continue
		}

		params := &chart.Params{
			Symbol:   symbol,
			Interval: "1wk",
		}
		quoteIter := chart.Get(params)
		for quoteIter.Next() {
			fmt.Println(quoteIter.Bar())
		}

		// contracts := optionCalculator.NewOptionCalculator(latestPrice, 10, iter)
		// putIV, err := contracts.GetATMPutIV()
		// if err != nil {
		// 	fmt.Println("fail to get put iv", symbol, dte)
		// 	continue
		// }
		// callIV, err := contracts.GetATMCallIV()
		// if err != nil {
		// 	fmt.Println("fail to get call iv", symbol, dte)
		// 	continue
		// }
		// putPremium, err := contracts.GetATMPutPremium()
		// if err != nil {
		// 	fmt.Println("fail to get put premium", symbol, dte)
		// 	continue
		// }
		// callPremium, err := contracts.GetATMCallPremium()
		// if err != nil {
		// 	fmt.Println("fail to get call premium", symbol, dte)
		// 	continue
		// }
		// fmt.Printf("%s: %.2f. DTE: %d. PutIV: %.2f. CallIV: %.2f. Put Premium %.2f. Call Premium %.2f.\n",
		// 	symbol, latestPrice, dte,
		// 	putIV, callIV,
		// 	putPremium, callPremium,
		// )
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}
