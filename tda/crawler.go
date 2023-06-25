package tda

import (
	"fmt"
	"math"
	"sort"
	"time"
	_ "time/tzdata"

	"github.com/tianhai82/ivsensor/market_data"
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
	to := today.AddDate(0, 0, 7)
	toDate := to.Format("2006-01-02")
	priceList, err := market_data.RetrieveOptions(s.Symbol, OptionContractPUT, OptionRangeOTM, toDate)
	if err != nil {
		return fmt.Errorf("fail to retrieve option chain: %v", err)
	}
	dte := priceList[0].Dte
	expDate := time.Unix(priceList[0].Expiration, 0)
	atrNormalized := s.WeeklyATR * math.Pow(numOfWeeks(int(dte)), 0.7)
	highestStrike := s.CurrentStockPrice - atrNormalized
	sort.Slice(priceList, func(i, j int) bool {
		return priceList[i].Strike < priceList[j].Strike
	})
	index := -1
	for i, price := range priceList {
		if price.Strike > highestStrike {
			index = i - 1
			break
		}
	}
	if index < 0 {
		fmt.Println(s.Symbol, "no suitable strike price found")
		return nil
	}
	if priceList[index].Strike > highestStrike {
		fmt.Println(s.Symbol, "no suitable strike price found")
		return nil
	}
	minSize := 1
	if priceList[index].Strike < 20 {
		minSize = 3
	} else if priceList[index].Strike < 50 {
		minSize = 2
	}
	if int(priceList[index].AskSize) < minSize || int(priceList[index].BidSize) < minSize {
		fmt.Println(s.Symbol, "bid or ask is empty")
		return nil
	}
	if (priceList[index].Ask / priceList[index].Bid) > 10.0 {
		return nil
	}
	premium := StockOptionPremium{
		ExpiryDate:    expDate,
		DTE:           int(dte),
		NormalizedATR: atrNormalized,
	}
	premium.PutStrike = priceList[index].Strike
	premium.PutPremium = (priceList[index].Bid + priceList[index].Ask) / 2
	premium.PutPremiumAnnualizedPercent = premium.PutPremium / premium.PutStrike / numOfWeeks(premium.DTE) * 52.0
	s.OptionPremiums = append(s.OptionPremiums, premium)

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

	// find the 60 percentile true range
	trueRange60, err := ta.TrueRangePercentileCandles(candles, 0.6)
	if err != nil {
		return err
	}

	// ATR is smaller than trueRange60, use trueRange60
	if atr < trueRange60 {
		atr = trueRange60
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
