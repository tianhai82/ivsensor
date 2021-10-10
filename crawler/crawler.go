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
	"github.com/plandem/xlsx"
	"github.com/tianhai82/ivsensor/firebase"
	"github.com/tianhai82/ivsensor/model"
	"github.com/tianhai82/ivsensor/optionCalculator"
	"github.com/tianhai82/ivsensor/ta"
	"github.com/tianhai82/ivsensor/telegram"
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
	saveOptionsRecords(dateStr)
	msg := fmt.Sprintf("Options analysis done for %s. Download excel report at https://api-zwv4vcvbqq-uc.a.run.app/doc/%s.xlsx", dateStr, dateStr)
	telegram.SendMessage(msg, "21450012", "1743013035:AAF43wU6BX4UOcHwL-vX2OGcM1xMhBoe0Ug")
	fmt.Println("done")
}

func saveOptionsRecords(today string) error {
	bucket, err := firebase.StorageClient.DefaultBucket()
	if err != nil {
		fmt.Println("fail to get bucket", err)
		return err
	}

	docIter := firebase.FirestoreClient.Collection("record").Where("Date", "==", today).Documents(context.Background())
	docs, err := docIter.GetAll()
	if err != nil {
		fmt.Println("fail to retrieve records from firestore", err)
		return err
	}

	excel := xlsx.New()
	sheet := excel.AddSheet("options")
	WriteHeader(sheet)
	row := 1
	for _, doc := range docs {
		var rec model.OptionRecord
		err = doc.DataTo(&rec)
		if err != nil {
			fmt.Println(err)
			continue
		}
		WriteRecord(sheet, row, rec)
		row++
	}

	filename := fmt.Sprintf("%s.xlsx", today)
	writer := bucket.Object(filename).NewWriter(context.Background())
	err = excel.SaveAs(writer)
	if err != nil {
		fmt.Println("fail to save excel to cloud storage", err)
		return err
	}
	err = writer.Close()
	if err != nil {
		fmt.Println("fail to close cloud storage writer", err)
		return err
	}
	// for _, doc := range docs {
	// 	doc.Ref.Delete(context.Background())
	// }
	return nil
}

func WriteRecord(sheet xlsx.Sheet, row int, rec model.OptionRecord) {
	sheet.Cell(0, row).SetText(rec.Symbol)
	sheet.Cell(1, row).SetFloat(rec.StockPrice)
	sheet.Cell(2, row).SetFloat(rec.NormalizedATR)
	sheet.Cell(3, row).SetFloat(rec.WeeklyATR)
	sheet.Cell(4, row).SetFloat(rec.WeeklyATRP)
	sheet.Cell(5, row).SetInt(rec.DTE)
	sheet.Cell(6, row).SetText(rec.ExpiryDate)
	sheet.Cell(7, row).SetFloat(rec.PutIVAtm)
	sheet.Cell(8, row).SetFloat(rec.CallIVAtm)
	sheet.Cell(9, row).SetFloat(rec.PutStrike)
	sheet.Cell(10, row).SetFloat(rec.PutPremium)
	sheet.Cell(11, row).SetFloat(rec.PutPremiumAnnualizedPercent)
	sheet.Cell(12, row).SetFloat(rec.CallStrike)
	sheet.Cell(13, row).SetFloat(rec.CallPremium)
	sheet.Cell(14, row).SetFloat(rec.CallPremiumAnnualizedPercent)
}

func WriteHeader(sheet xlsx.Sheet) {
	sheet.Cell(0, 0).SetText("Symbol")
	sheet.Cell(1, 0).SetText("Stock Price")
	sheet.Cell(2, 0).SetText("Normalized ATR")
	sheet.Cell(3, 0).SetText("Weekly ATR")
	sheet.Cell(4, 0).SetText("Weekly ATRP")
	sheet.Cell(5, 0).SetText("DTE")
	sheet.Cell(6, 0).SetText("Expiry Date")
	sheet.Cell(7, 0).SetText("Put IV ATM")
	sheet.Cell(8, 0).SetText("Call IV ATM")
	sheet.Cell(9, 0).SetText("Put Strike")
	sheet.Cell(10, 0).SetText("Put Premium")
	sheet.Cell(11, 0).SetText("Put Premium Annualized %")
	sheet.Cell(12, 0).SetText("Call Strike")
	sheet.Cell(13, 0).SetText("Call Premium")
	sheet.Cell(14, 0).SetText("Call Premium Annualized %")
}

func CrawlSymbol(symbol string) ([]model.OptionRecord, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in CrawlSymbol", r)
		}
	}()
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

	// find the 60 percentile true range
	trueRange60, err := ta.TrueRangePercentile(bars, 0.6)
	if err != nil {
		return nil, err
	}

	// ATR is smaller than trueRange60, use trueRange60
	if atr < trueRange60 {
		atr = trueRange60
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

		atrNormalized := atr * math.Pow(optionCalculator.NumOfWeeks(dte), 0.7)

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
