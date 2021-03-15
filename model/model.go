package model

type Stock struct {
	TotalVolume    int    `json:"TotalVolume"`
	Symbol         string `json:"Symbol"`
	CompanyName    string `json:"CompanyName"`
	Sector         string `json:"Sector"`
	CallVolume     int    `json:"CallVolume"`
	PutVolume      int    `json:"PutVolume"`
	AvgVolume90Day int    `json:"AvgVolume_90Day"`
	RelativeVolume int    `json:"RelativeVolume"`
	TradeCount     int    `json:"TradeCount"`
	PutPercentage  int    `json:"PutPercentage"`
}

type DayTask struct {
	SymbolsStatuses map[string]bool `firestore:"symbolsStatuses"`
	ID              string          `firestore:"id"`
}

type OptionRecord struct {
	Symbol                       string
	StockPrice                   float64
	NormalizedATR                float64
	WeeklyATR                    float64
	WeeklyATRP                   float64
	DTE                          int
	ExpiryDate                   string
	PutIVAtm                     float64
	CallIVAtm                    float64
	PutStrike                    float64
	PutPremium                   float64
	PutPremiumAnnualizedPercent  float64
	CallStrike                   float64
	CallPremium                  float64
	CallPremiumAnnualizedPercent float64
}
