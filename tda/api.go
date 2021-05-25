package tda

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/plandem/xlsx"
	"github.com/tianhai82/ivsensor/firebase"
	"github.com/tianhai82/ivsensor/model"
	"github.com/tianhai82/ivsensor/telegram"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func AddApi(router *gin.RouterGroup) {
	router.GET("/options", crawlOptions)
}

func crawlOptions(c *gin.Context) {
	startTime := time.Now()
	dateStr := time.Now().Format("2006-01-02")
	task, err := firebase.FirestoreClient.Collection("tdaTask").Doc(dateStr).Get(context.Background())
	dayTask := model.DayTask{
		ID:              dateStr,
		SymbolsStatuses: nil,
	}
	if err != nil {
		if status.Code(err) == codes.NotFound {
			fmt.Println("day task not found. creating")
			statuses := map[string]bool{}
			for _, s := range firebase.StockSymbols {
				statuses[s] = false
			}
			dayTask.SymbolsStatuses = statuses
			firebase.FirestoreClient.Collection("tdaTask").Doc(dateStr).Set(context.Background(), dayTask)
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
			stockAtr := StockATR{
				Symbol: symbol,
			}

			err = stockAtr.PopulateATR(dateStr)
			time.Sleep(50 * time.Millisecond)
			if err != nil {
				fmt.Println(symbol, err)
			} else {
				err2 := stockAtr.RetrieveOptionPremium()
				if err2 != nil {
					fmt.Println(symbol, err)
				} else {
					if len(stockAtr.OptionPremiums) > 0 {
						firebase.FirestoreClient.Collection("tdaRecord").
							Doc(fmt.Sprintf("%s|%s", stockAtr.Symbol, stockAtr.CurrentDate)).
							Set(context.Background(), stockAtr)
						// firebase.FirestoreClient.Collection("tdaRecord").Add(context.Background(), stockAtr)
					}
				}
			}

			firebase.FirestoreClient.Collection("tdaTask").Doc(dateStr).Update(context.Background(),
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

	docIter := firebase.FirestoreClient.Collection("tdaRecord").Where("CurrentDate", "==", today).Documents(context.Background())
	docs, err := docIter.GetAll()
	if err != nil {
		fmt.Println("fail to retrieve records from firestore", err)
		return err
	}

	atrs := make([]StockATR, 0, len(docs))
	for _, doc := range docs {
		var rec StockATR
		err = doc.DataTo(&rec)
		if err != nil {
			fmt.Println(err)
			continue
		}
		atrs = append(atrs, rec)
	}

	excel := xlsx.New()
	sheet := excel.AddSheet("options")
	writeHeader(sheet)
	row := 1
	for _, rec := range atrs {
		bestStocksMap := findBestPair(rec, atrs)
		for _, premium := range rec.OptionPremiums {
			writeRecord(sheet, row, rec, premium, bestStocksMap)
			row++
		}

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
	return nil
}

func findBestPair(atr StockATR, atrs []StockATR) map[float64]StockATR {
	var scores []float64
	scoresAtrMap := make(map[float64]StockATR)
	for _, a := range atrs {
		score, err := PairScore(atr.Candles, a.Candles)
		if err != nil || score >= 0.5 {
			continue
		}
		scores = append(scores, score)
		scoresAtrMap[score] = a
	}

	sort.Sort(sort.Reverse(sort.Float64Slice(scores)))
	outMap := make(map[float64]StockATR)

	for i, score := range scores {
		outMap[score] = scoresAtrMap[score]
		if i == 2 {
			break
		}
	}
	return outMap
}

func PairScore(candles1, candles2 []model.Candle) (float64, error) {
	if len(candles1) != len(candles2) {
		return 0.0, fmt.Errorf("different slice lenght. 1st: %d. 2nd: %d.", len(candles1), len(candles2))
	}
	score := 0.0
	for i, c := range candles1 {
		var x, y int
		if c.Close > c.Open {
			x = 1
		}
		if candles2[i].Close > candles2[i].Open {
			y = 1
		}
		if x != y {
			score += 1.0
		}
	}
	return score / float64(len(candles1)), nil
}

// func findMaxNegativeCorr(atr StockATR, atrs []StockATR) (map[float64]StockATR, error) {
// 	var corrs []float64
// 	corrAtrMap := make(map[float64]StockATR)
// 	for _, a := range atrs {
// 		corr, err := stats.Correlation(atr.Closes, a.Closes)
// 		if err != nil || corr >= 0 {
// 			continue
// 		}
// 		corrs = append(corrs, corr)
// 		corrAtrMap[corr] = a
// 	}
// 	if len(corrs) == 0 {
// 		return nil, fmt.Errorf("no negative correlation found")
// 	}
// 	sort.Float64s(corrs)
// 	outMap := make(map[float64]StockATR)
// 	for i := 0; i < 3; i++ {
// 		corr := corrs[i]
// 		outMap[corr] = corrAtrMap[corr]
// 	}
// 	return outMap, nil
// }

func writeRecord(sheet xlsx.Sheet, row int, rec StockATR, premium StockOptionPremium, pairMap map[float64]StockATR) {
	sheet.Cell(0, row).SetText(rec.Symbol)
	sheet.Cell(1, row).SetFloat(rec.CurrentStockPrice)
	sheet.Cell(2, row).SetFloat(rec.WeeklyATR)

	sheet.Cell(3, row).SetFloat(premium.NormalizedATR)
	sheet.Cell(4, row).SetInt(premium.DTE)
	sheet.Cell(5, row).SetDate(premium.ExpiryDate)

	sheet.Cell(6, row).SetFloat(premium.PutStrike)
	sheet.Cell(7, row).SetFloat(premium.PutPremium)
	sheet.Cell(8, row).SetFloat(premium.PutPremiumAnnualizedPercent)

	col := 9
	for corr, atr := range pairMap {
		sheet.Cell(col, row).SetText(atr.Symbol)
		col += 1
		sheet.Cell(col, row).SetFloat(corr)
		col += 1
		sheet.Cell(col, row).SetFloat(atr.CurrentStockPrice)
		col += 1
	}
}

func writeHeader(sheet xlsx.Sheet) {
	sheet.Cell(0, 0).SetText("Symbol")
	sheet.Cell(1, 0).SetText("Stock Price")
	sheet.Cell(2, 0).SetText("Weekly ATR")

	sheet.Cell(3, 0).SetText("Normalized ATR")
	sheet.Cell(4, 0).SetText("DTE")
	sheet.Cell(5, 0).SetText("Expiry Date")

	sheet.Cell(6, 0).SetText("Put Strike")
	sheet.Cell(7, 0).SetText("Put Premium")
	sheet.Cell(8, 0).SetText("Put Premium Annualized %")

	sheet.Cell(9, 0).SetText("Most Negative Corr Symbol")
	sheet.Cell(10, 0).SetText("Correlation")
	sheet.Cell(11, 0).SetText("Stock Price")
	sheet.Cell(12, 0).SetText("Most Negative Corr Symbol")
	sheet.Cell(13, 0).SetText("Correlation")
	sheet.Cell(14, 0).SetText("Stock Price")
	sheet.Cell(15, 0).SetText("Most Negative Corr Symbol")
	sheet.Cell(16, 0).SetText("Correlation")
	sheet.Cell(17, 0).SetText("Stock Price")

}
