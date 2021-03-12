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
