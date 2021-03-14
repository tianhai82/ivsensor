package crawler

import (
	"context"
	"fmt"
	"math"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/chart"
	"github.com/piquette/finance-go/datetime"
	"github.com/piquette/finance-go/options"
	"github.com/piquette/finance-go/quote"
	"github.com/tianhai82/ivsensor/firebase"
	"github.com/tianhai82/ivsensor/model"
	"github.com/tianhai82/ivsensor/optionCalculator"
	"github.com/tianhai82/ivsensor/ta"
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
			CrawlSymbol(symbol)
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

func CrawlSymbol(symbol string) error {
	println("crawling", symbol)

	q, err := quote.Get(symbol)
	if err != nil {
		fmt.Println("fail to get quote for", symbol, err)
	}
	latestPrice := q.RegularMarketPrice

	straddle := options.GetStraddle(symbol)
	meta := straddle.Meta()
	now := time.Now()
	for _, d := range meta.AllExpirationDates {
		dt := datetime.FromUnix(d)
		days := dt.Time().Sub(now).Hours() / 24
		dte := int(math.Ceil(days))
		if dte > 70 || dte <= 0 {
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

		start := time.Now()
		start = start.AddDate(0, -3, 0)
		end := time.Now()
		params := &chart.Params{
			Symbol:   symbol,
			Interval: "1wk",
			Start:    datetime.New(&start),
			End:      datetime.New(&end),
		}
		quoteIter := chart.Get(params)
		var bars []finance.ChartBar
		for quoteIter.Next() {
			bars = append(bars, *quoteIter.Bar())
		}
		atr, err := ta.ATR(bars, 4)
		if err != nil {
			fmt.Println(err)
			continue
		}
		atrp, err := ta.ATRP(bars, 4)
		if err != nil {
			fmt.Println(err)
			continue
		}

		optCalc := optionCalculator.NewOptionCalculator(latestPrice, atr, dte, iter)
		putIV, err := optCalc.GetATMPutIV()
		if err != nil {
			fmt.Println("fail to get put iv", symbol, dte)
			continue
		}
		callIV, err := optCalc.GetATMCallIV()
		if err != nil {
			fmt.Println("fail to get call iv", symbol, dte)
			continue
		}
		putStrike, putPremium, putPremiumPercentAnnual, err := optCalc.GetPutPremium()
		if err != nil {
			fmt.Println("fail to get put premium", symbol, dte)

		}
		callStrike, callPremium, callPremiumAnnual, err := optCalc.GetCallPremium()
		if err != nil {
			fmt.Println("fail to get call premium", symbol, dte)

		}
		fmt.Printf("%s: %.2f. DTE: %d. PutIV: %.2f. CallIV: %.2f. Put strike %.2f. PutPremium %.2f. Put Premium Percent %.2f. Call strike %.2f. CallPremium %.2f. Call Premium Percent %.2f. ATRP: %.2f \n",
			symbol, latestPrice, dte,
			putIV, callIV,
			putStrike, putPremium, putPremiumPercentAnnual,
			callStrike, callPremium, callPremiumAnnual,
			atrp,
		)
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}
