package tda

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/montanaflynn/stats"
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
			if err != nil {
				fmt.Println(symbol, err)
			} else {
				err2 := stockAtr.RetrieveOptionPremium()
				if err2 != nil {
					fmt.Println(symbol, err)
				} else {
					if len(stockAtr.OptionPremiums) > 0 {
						firebase.FirestoreClient.Collection("tdaRecord").Add(context.Background(), stockAtr)
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
			time.Sleep(500 * time.Millisecond)
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
		negAtr, corr, _ := findMaxNegativeCorr(rec, atrs)
		for _, premium := range rec.OptionPremiums {
			writeRecord(sheet, row, rec, premium, negAtr, corr)
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

func findMaxNegativeCorr(atr StockATR, atrs []StockATR) (StockATR, float64, error) {
	smallest := 1.0
	var temp StockATR
	for _, a := range atrs {
		corr, err := stats.Correlation(atr.Closes, a.Closes)
		if err == nil && corr < smallest {
			smallest = corr
			temp = a
		}
	}
	if smallest < 0 {
		return temp, smallest, nil
	}
	return StockATR{}, 0.0, fmt.Errorf("no negative found")
}

func writeRecord(sheet xlsx.Sheet, row int, rec StockATR, premium StockOptionPremium, negAtr StockATR, corr float64) {
	sheet.Cell(0, row).SetText(rec.Symbol)
	sheet.Cell(1, row).SetFloat(rec.CurrentStockPrice)
	sheet.Cell(2, row).SetFloat(rec.WeeklyATR)

	sheet.Cell(3, row).SetFloat(premium.NormalizedATR)
	sheet.Cell(4, row).SetInt(premium.DTE)
	sheet.Cell(5, row).SetText(premium.ExpiryDate)

	sheet.Cell(6, row).SetFloat(premium.PutStrike)
	sheet.Cell(7, row).SetFloat(premium.PutPremium)
	sheet.Cell(8, row).SetFloat(premium.PutPremiumAnnualizedPercent)

	if corr != 0.0 {
		sheet.Cell(9, row).SetText(negAtr.Symbol)
		sheet.Cell(10, row).SetFloat(corr)
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

}
