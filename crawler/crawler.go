package crawler

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
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
	startTime := time.Now()
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
	total := len(dayTask.SymbolsStatuses)

	for symbol, status := range dayTask.SymbolsStatuses {
		i++
		if !status {
			fmt.Println("processing", symbol)
			records, err := CrawlSymbol(symbol)
			if err != nil {
				fmt.Println("cannot crawl", symbol, err)
			}

			for _, rec := range records {
				firebase.FirestoreClient.Collection("record").Add(context.Background(), rec)
			}

			firebase.FirestoreClient.Collection("task").Doc(dateStr).Update(context.Background(),
				[]firestore.Update{
					{
						FieldPath: firestore.FieldPath{"symbolsStatuses", symbol},
						Value:     true,
					},
				},
			)

			if i%10 == 0 {
				fmt.Printf("%d out of %d.\n", i, total)
			}
		}
		duration := time.Since(startTime)
		if duration.Minutes() > 55 {
			c.AbortWithError(http.StatusRequestTimeout, fmt.Errorf("taking too long"))
			return
		}
	}
	fmt.Println("done")
}

func CrawlSymbol(symbol string) ([]model.OptionRecord, error) {
	println("crawling", symbol)

	q, err := quote.Get(symbol)
	if err != nil || q == nil {
		fmt.Println("fail to get quote for", symbol, err)
		return nil, fmt.Errorf("fail to get quote for %s", symbol)
	}
	latestPrice := q.RegularMarketPrice
	if latestPrice > 150 {
		return nil, fmt.Errorf("ignoring high priced stocks")
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
		return nil, err
	}
	atrp, err := ta.ATRP(bars, 4)
	if err != nil {
		return nil, err
	}

	straddle := options.GetStraddle(symbol)
	if straddle == nil || straddle.Count() == 0 {
		return nil, errors.New("no straddle found")
	}
	meta := straddle.Meta()
	if meta == nil {
		return nil, errors.New("no straddle found")
	}
	now := time.Now()
	records := make([]model.OptionRecord, 0)
	for _, d := range meta.AllExpirationDates {
		dt := datetime.FromUnix(d)
		days := dt.Time().Sub(now).Hours() / 24
		dte := int(math.Ceil(days + 1))
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

		optCalc := optionCalculator.NewOptionCalculator(latestPrice, atr, dte, iter)
		putIV, _ := optCalc.GetATMPutIV()

		callIV, _ := optCalc.GetATMCallIV()

		putStrike, putPremium, putPremiumPercentAnnual, _ := optCalc.GetPutPremium()

		callStrike, callPremium, callPremiumAnnual, _ := optCalc.GetCallPremium()

		atrNormalized := atr * math.Pow((float64(dte)/7.0), 0.75)

		now := time.Now().UTC()
		rec := model.OptionRecord{
			Date:                         now.Format("2006-01-02"),
			Symbol:                       symbol,
			StockPrice:                   latestPrice,
			NormalizedATR:                atrNormalized,
			WeeklyATR:                    atr,
			WeeklyATRP:                   atrp,
			DTE:                          dte,
			ExpiryDate:                   dt.Time().Format("2006-01-02"),
			PutIVAtm:                     putIV,
			CallIVAtm:                    callIV,
			PutStrike:                    putStrike,
			PutPremium:                   putPremium,
			PutPremiumAnnualizedPercent:  putPremiumPercentAnnual,
			CallStrike:                   callStrike,
			CallPremium:                  callPremium,
			CallPremiumAnnualizedPercent: callPremiumAnnual,
		}
		if rec.PutPremiumAnnualizedPercent > 0.30 || rec.CallPremiumAnnualizedPercent > 0.3 {
			records = append(records, rec)
		}
		time.Sleep(100 * time.Millisecond)
	}
	return records, nil
}
