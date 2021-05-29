package model

import "time"

type Candle struct {
	Symbol   string    `json:"Symbol"`
	FromDate time.Time `json:"FromDate"`
	Open     float64   `json:"Open"`
	High     float64   `json:"High"`
	Low      float64   `json:"Low"`
	Close    float64   `json:"Close"`
}

type Stock struct {
	TotalVolume    int     `json:"TotalVolume"`
	Symbol         string  `json:"Symbol"`
	CompanyName    string  `json:"CompanyName"`
	Sector         string  `json:"Sector"`
	CallVolume     int     `json:"CallVolume"`
	PutVolume      int     `json:"PutVolume"`
	AvgVolume90Day int     `json:"AvgVolume_90Day"`
	RelativeVolume float64 `json:"RelativeVolume"`
	TradeCount     int     `json:"TradeCount"`
	PutPercentage  float64 `json:"PutPercentage"`
}

type DayTask struct {
	SymbolsStatuses map[string]bool `firestore:"symbolsStatuses"`
	ID              string          `firestore:"id"`
}

type OptionRecord struct {
	Date                         string
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

type Chains struct {
	Symbol            string     `json:"symbol"`
	Status            string     `json:"status"`
	Underlying        Underlying `json:"underlying"`
	Strategy          string     `json:"strategy"`
	Interval          float64    `json:"interval"`
	IsDelayed         bool       `json:"isDelayed"`
	IsIndex           bool       `json:"isIndex"`
	InterestRate      float64    `json:"interestRate"`
	UnderlyingPrice   float64    `json:"underlyingPrice"`
	Volatility        float64    `json:"volatility"`
	DaysToExpiration  float64    `json:"daysToExpiration"`
	NumberOfContracts int        `json:"numberOfContracts"`
	CallExpDateMap    ExpDateMap `json:"callExpDateMap"`
	PutExpDateMap     ExpDateMap `json:"putExpDateMap"`
}
type ExpDateMap map[string]map[string][]ExpDateOption
type ExpDateOption struct {
	PutCall                string      `json:"putCall"`
	Symbol                 string      `json:"symbol"`
	Description            string      `json:"description"`
	ExchangeName           string      `json:"exchangeName"`
	Bid                    float64     `json:"bid"`
	Ask                    float64     `json:"ask"`
	Last                   float64     `json:"last"`
	Mark                   float64     `json:"mark"`
	BidSize                int         `json:"bidSize"`
	AskSize                int         `json:"askSize"`
	BidAskSize             string      `json:"bidAskSize"`
	LastSize               float64     `json:"lastSize"`
	HighPrice              float64     `json:"highPrice"`
	LowPrice               float64     `json:"lowPrice"`
	OpenPrice              float64     `json:"openPrice"`
	ClosePrice             float64     `json:"closePrice"`
	TotalVolume            int         `json:"totalVolume"`
	TradeDate              string      `json:"tradeDate"`
	TradeTimeInLong        int         `json:"tradeTimeInLong"`
	QuoteTimeInLong        int         `json:"quoteTimeInLong"`
	NetChange              float64     `json:"netChange"`
	Volatility             interface{} `json:"volatility"`
	Delta                  interface{} `json:"delta"`
	Gamma                  interface{} `json:"gamma"`
	Theta                  interface{} `json:"theta"`
	Vega                   interface{} `json:"vega"`
	Rho                    interface{} `json:"rho"`
	OpenInterest           int         `json:"openInterest"`
	TimeValue              interface{} `json:"timeValue"`
	TheoreticalOptionValue interface{} `json:"theoreticalOptionValue"`
	TheoreticalVolatility  interface{} `json:"theoreticalVolatility"`
	OptionDeliverablesList string      `json:"optionDeliverablesList"`
	StrikePrice            float64     `json:"strikePrice"`
	ExpirationDate         int         `json:"expirationDate"`
	DaysToExpiration       int         `json:"daysToExpiration"`
	ExpirationType         string      `json:"expirationType"`
	LastTradingDate        int         `json:"lastTradingDay"`
	Multiplier             float64     `json:"multiplier"`
	SettlementType         string      `json:"settlementType"`
	DeliverableNote        string      `json:"deliverableNote"`
	IsIndexOption          bool        `json:"isIndexOption"`
	PercentChange          float64     `json:"percentChange"`
	MarkChange             float64     `json:"markChange"`
	MarkPercentChange      float64     `json:"markPercentChange"`
	InTheMoney             bool        `json:"inTheMoney"`
	Mini                   bool        `json:"mini"`
	NonStandard            bool        `json:"nonStandard"`
}
type Underlying struct {
	Symbol            string  `json:"symbol"`
	Description       string  `json:"description"`
	Change            float64 `json:"change"`
	PercentChange     float64 `json:"percentChange"`
	Close             float64 `json:"close"`
	QuoteTime         int     `json:"quoteTime"`
	TradeTime         int     `json:"tradeTime"`
	Bid               float64 `json:"bid"`
	Ask               float64 `json:"ask"`
	Last              float64 `json:"last"`
	Mark              float64 `json:"mark"`
	MarkChange        float64 `json:"markChange"`
	MarkPercentChange float64 `json:"markPercentChange"`
	BidSize           int     `json:"bidSize"`
	AskSize           int     `json:"askSize"`
	HighPrice         float64 `json:"highPrice"`
	LowPrice          float64 `json:"lowPrice"`
	OpenPrice         float64 `json:"openPrice"`
	TotalVolume       int     `json:"totalVolume"`
	ExchangeName      string  `json:"exchangeName"`
	FiftyTwoWeekHigh  float64 `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow   float64 `json:"fiftyTwoWeekLow"`
	Delayed           bool    `json:"delayed"`
}
