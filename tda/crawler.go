package tda

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/tianhai82/ivsensor/model"
	"github.com/tianhai82/ivsensor/ta"
)

type StockOptionPremium struct {
	ExpiryDate                  time.Time
	DTE                         int
	NormalizedATR               float64
	PutStrike                   float64
	PutPremium                  float64
	PutPremiumAnnualizedPercent float64
}

type StockATR struct {
	Symbol            string
	WeeklyATR         float64
	CurrentDate       string
	CurrentStockPrice float64
	OptionPremiums    []StockOptionPremium
	Candles           []model.Candle `json:"-"`
}

func (s *StockATR) RetrieveOptionPremium() error {
	if s.WeeklyATR == 0.0 {
		return fmt.Errorf("weekly ATR must be populate first")
	}

	utc, _ := time.LoadLocation("UTC")
	var today time.Time
	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		zone = utc
	}
	today = time.Now().In(zone)

	fromDate := today.Format("2006-01-02")
	to := today.AddDate(0, 0, 14)
	toDate := to.Format("2006-01-02")
	chains, err := RetrieveOptions(s.Symbol, OptionContractPUT, OptionRangeOTM, fromDate, toDate)
	if err != nil {
		return fmt.Errorf("fail to retrieve option chain: %v", err)
	}
	if chains.Status != "SUCCESS" {
		return fmt.Errorf("api status not success")
	}

	for expirationDate, priceMap := range chains.PutExpDateMap {
		segment := strings.Split(expirationDate, ":")
		if len(segment) != 2 {
			fmt.Println(s.Symbol, "invalid expirationDate found", expirationDate)
			continue
		}
		expDate, err := time.ParseInLocation("2006-01-02", segment[0], zone)
		if err != nil {
			fmt.Println(s.Symbol, "invalid expirationDate found", expirationDate)
			continue
		}
		dte, err := strconv.Atoi(segment[1])
		if err != nil {
			fmt.Println(s.Symbol, "invalid expirationDate found", expirationDate)
			continue
		}
		atrNormalized := s.WeeklyATR * math.Pow(numOfWeeks(dte), 0.7)

		highestStrike := s.CurrentStockPrice - atrNormalized
		premium := StockOptionPremium{
			ExpiryDate:    expDate,
			DTE:           dte,
			NormalizedATR: atrNormalized,
		}

		var priceList []model.ExpDateOption
		for _, list := range priceMap {
			if len(list) < 1 {
				continue
			}
			priceList = append(priceList, list[0])
		}
		sort.Slice(priceList, func(i, j int) bool {
			return priceList[i].StrikePrice < priceList[j].StrikePrice
		})
		index := -1
		for i, price := range priceList {
			if price.StrikePrice > highestStrike {
				index = i - 1
				break
			}
		}
		if index < 0 {
			fmt.Println(s.Symbol, "no suitable strike price found")
			continue
		}
		if priceList[index].StrikePrice > highestStrike {
			fmt.Println(s.Symbol, "no suitable strike price found")
			continue
		}
		minSize := 1
		if priceList[index].StrikePrice < 20 {
			minSize = 3
		} else if priceList[index].StrikePrice < 50 {
			minSize = 2
		}
		if priceList[index].AskSize < minSize || priceList[index].BidSize < minSize {
			fmt.Println(s.Symbol, "bid or ask is empty")
			continue
		}
		if (priceList[index].Ask / priceList[index].Bid) > 10.0 {
			continue
		}
		premium.PutStrike = priceList[index].StrikePrice
		premium.PutPremium = (priceList[index].Bid + priceList[index].Ask) / 2
		premium.PutPremiumAnnualizedPercent = premium.PutPremium / premium.PutStrike / numOfWeeks(premium.DTE) * 52.0
		s.OptionPremiums = append(s.OptionPremiums, premium)
	}

	return nil
}

func (s *StockATR) PopulateATR(date string) error {
	candles, err := RetrieveHistory(s.Symbol, FrequencyWeekly, 13)
	if err != nil {
		return fmt.Errorf("fail to retrieve weekly stock history: %v", err)
	}
	s.CurrentStockPrice = candles[len(candles)-1].Close

	var now time.Time
	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		fmt.Println(err)
		now = time.Now()
	} else {
		now = time.Now().In(zone)
	}

	dayOfWeek := now.Weekday()
	if dayOfWeek == time.Monday || dayOfWeek == time.Tuesday || dayOfWeek == time.Wednesday {
		candles = candles[:len(candles)-1]
	}
	s.Candles = candles
	atr, err := ta.ATRCandles(candles, 4)
	if err != nil {
		return err
	}
	s.WeeklyATR = atr
	s.CurrentDate = date
	return nil
}

func numOfWeeks(dte int) float64 {
	if dte <= 7 && dte >= 5 {
		return 1.0
	}
	return float64(dte/7) + float64(dte%7)/5.0
}
