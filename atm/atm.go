package atm

import (
	"fmt"

	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/options"
)

type ATM struct {
	prevCall    *finance.Contract
	prevPut     *finance.Contract
	currentCall *finance.Contract
	currentPut  *finance.Contract
}

func NewATM(price float64, iter *options.StraddleIter) *ATM {
	return findATMContracts(price, iter)
}

func (contracts *ATM) GetATMPutIV() (float64, error) {
	if contracts.prevPut == nil || contracts.currentPut == nil {
		return 0.0, fmt.Errorf("missing contract")
	}
	return (contracts.prevPut.ImpliedVolatility + contracts.currentPut.ImpliedVolatility) / 2, nil
}
func (contracts *ATM) GetATMCallIV() (float64, error) {
	if contracts.prevCall == nil || contracts.currentCall == nil {
		return 0.0, fmt.Errorf("missing contract")
	}
	return (contracts.prevCall.ImpliedVolatility + contracts.currentCall.ImpliedVolatility) / 2, nil
}
func (contracts *ATM) GetATMPutPremium() (float64, error) {
	if contracts.prevPut == nil || contracts.currentPut == nil {
		return 0.0, fmt.Errorf("missing contract")
	}
	return (contracts.prevPut.Ask + contracts.currentPut.Ask + contracts.prevPut.Bid + contracts.currentPut.Bid) / 4, nil
}
func (contracts *ATM) GetATMCallPremium() (float64, error) {
	if contracts.prevCall == nil || contracts.currentCall == nil {
		return 0.0, fmt.Errorf("missing contract")
	}
	return (contracts.prevCall.Ask + contracts.currentCall.Ask + contracts.prevCall.Bid + contracts.currentCall.Bid) / 4, nil
}

func findATMContracts(currentPrice float64, iter *options.StraddleIter) *ATM {
	var prevCall *finance.Contract
	var prevPut *finance.Contract
	var currentCall *finance.Contract
	var currentPut *finance.Contract
	for iter.Next() {
		stra := iter.Straddle()
		if stra.Strike > currentPrice {
			currentCall = stra.Call
			currentPut = stra.Put
			break
		} else {
			prevCall = stra.Call
			prevPut = stra.Put
		}
	}
	return &ATM{
		prevCall:    prevCall,
		prevPut:     prevPut,
		currentCall: currentCall,
		currentPut:  currentPut,
	}
}
