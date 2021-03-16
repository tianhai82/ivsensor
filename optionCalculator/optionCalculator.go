package optionCalculator

import (
	"fmt"
	"math"

	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/options"
)

type OptionCalculator struct {
	UnderlyingPrice     float64
	UnderLyingWeeklyATR float64
	DTE                 int
	OptionsStraddleIter *options.StraddleIter
	*atm
	putContract  *finance.Contract
	callContract *finance.Contract
}
type atm struct {
	prevCall    *finance.Contract
	prevPut     *finance.Contract
	currentCall *finance.Contract
	currentPut  *finance.Contract
}

func NewOptionCalculator(price, weeklyAtr float64, dte int, iter *options.StraddleIter) *OptionCalculator {
	straddles := toSlice(iter)
	a := findATMContracts(price, straddles)
	p := findPutContract(price, weeklyAtr, dte, straddles)
	c := findCallContract(price, weeklyAtr, dte, straddles)
	return &OptionCalculator{
		UnderlyingPrice:     price,
		UnderLyingWeeklyATR: weeklyAtr,
		OptionsStraddleIter: iter,
		DTE:                 dte,
		atm:                 a,
		putContract:         p,
		callContract:        c,
	}
}

func (optCalc *OptionCalculator) GetATMPutIV() (float64, error) {
	if optCalc.prevPut == nil || optCalc.currentPut == nil {
		return 0.0, fmt.Errorf("missing contract")
	}
	return (optCalc.prevPut.ImpliedVolatility + optCalc.currentPut.ImpliedVolatility) / 2, nil
}
func (optCalc *OptionCalculator) GetATMCallIV() (float64, error) {
	if optCalc.prevCall == nil || optCalc.currentCall == nil {
		return 0.0, fmt.Errorf("missing contract")
	}
	return (optCalc.prevCall.ImpliedVolatility + optCalc.currentCall.ImpliedVolatility) / 2, nil
}
func (optCalc *OptionCalculator) GetPutPremium() (float64, float64, float64, error) {
	if optCalc.putContract == nil {
		return 0.0, 0.0, 0.0, fmt.Errorf("missing contract")
	}
	if optCalc.putContract.Volume < 5 || optCalc.putContract.OpenInterest < 20 {
		return 0.0, 0.0, 0.0, fmt.Errorf("volume too low")
	}
	premium := (optCalc.putContract.Ask + optCalc.putContract.Bid) / 2
	premiumAnnualised := premium / optCalc.putContract.Strike / float64(optCalc.DTE) * 365.0
	return optCalc.putContract.Strike, premium, premiumAnnualised, nil
}
func (optCalc *OptionCalculator) GetCallPremium() (float64, float64, float64, error) {
	if optCalc.callContract == nil {
		return 0.0, 0.0, 0.0, fmt.Errorf("missing contract")
	}
	if optCalc.callContract.Volume < 5 || optCalc.callContract.OpenInterest < 20 {
		return 0.0, 0.0, 0.0, fmt.Errorf("volume too low")
	}
	premium := (optCalc.callContract.Ask + optCalc.callContract.Bid) / 2
	premiumAnnualised := premium / optCalc.callContract.Strike / float64(optCalc.DTE) * 365.0
	return optCalc.callContract.Strike, premium, premiumAnnualised, nil
}

func toSlice(iter *options.StraddleIter) []*finance.Straddle {
	var straddles []*finance.Straddle
	for iter.Next() {
		stra := iter.Straddle()
		straddles = append(straddles, stra)
	}
	return straddles
}

func findATMContracts(currentPrice float64, straddles []*finance.Straddle) *atm {
	var prevCall *finance.Contract
	var prevPut *finance.Contract
	var currentCall *finance.Contract
	var currentPut *finance.Contract
	for _, straddle := range straddles {
		if straddle.Strike > currentPrice {
			currentCall = straddle.Call
			currentPut = straddle.Put
			break
		} else {
			prevCall = straddle.Call
			prevPut = straddle.Put
		}
	}
	return &atm{
		prevCall:    prevCall,
		prevPut:     prevPut,
		currentCall: currentCall,
		currentPut:  currentPut,
	}
}

func findPutContract(price, weeklyAtr float64, dte int, straddles []*finance.Straddle) *finance.Contract {
	atr := weeklyAtr * math.Pow((float64(dte)/7.0), 0.75)
	s := price - atr
	var prevPut *finance.Contract
	for _, straddle := range straddles {
		if straddle.Strike > s {
			break
		} else {
			prevPut = straddle.Put
		}
	}
	return prevPut
}
func findCallContract(price, weeklyAtr float64, dte int, straddles []*finance.Straddle) *finance.Contract {
	atr := weeklyAtr * math.Pow((float64(dte)/7.0), 0.75)
	s := price + atr
	var currentCall *finance.Contract
	for _, straddle := range straddles {
		if straddle.Strike > s {
			currentCall = straddle.Call
			break
		}
	}
	return currentCall
}
