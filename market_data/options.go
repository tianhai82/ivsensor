package market_data

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/tianhai82/ivsensor/model"
)

const baseURL = "https://api.marketdata.app/v1/options/chain/%s/?expiration=%s&side=%s&range=%s&maxBidAskSpreadPct=0.5"

func RetrieveOptions(symbol, contractType, optionRange, expirationDate string) ([]model.StrikeDetails, error) {
	url := fmt.Sprintf(baseURL, symbol, expirationDate, contractType, optionRange)
	var resp model.MarketDataResp
	err := makeGetRequest(url, &resp)
	if err != nil {
		return nil, fmt.Errorf("RetrieveOptions http request failed: %w", err)
	}
	if resp.S != "ok" {
		return nil, fmt.Errorf("not ok")
	}
	if len(resp.Strike) <= 0 {
		return nil, fmt.Errorf("nothing returned")
	}
	out := make([]model.StrikeDetails, 0, len(resp.Strike))
	for i, strike := range resp.Strike {
		strikeDetail := model.StrikeDetails{
			Strike:       strike,
			Ask:          resp.Ask[i],
			AskSize:      resp.AskSize[i],
			Bid:          resp.Bid[i],
			BidSize:      resp.BidSize[i],
			Dte:          resp.Dte[i],
			Expiration:   resp.Expiration[i],
			Mid:          resp.Mid[i],
			OpenInterest: resp.OpenInterest[i],
			Side:         resp.Side[i],
		}
		out = append(out, strikeDetail)
	}
	return out, nil
}

func makeGetRequest(urlStr string, output interface{}) (err error) {
	resp, err := makeRequest(urlStr)
	if err != nil {
		err = errors.Wrap(err, "http get fails")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == 403 {
			resp, err = makeRequest(urlStr)
			if err != nil {
				err = errors.Wrap(err, "http get fails")
				return
			}
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return errors.New(resp.Status)
			}
		} else {
			return errors.New(resp.Status)
		}
	}

	var reader io.ReadCloser
	respEncoding := resp.Header.Get("Content-Encoding")
	switch respEncoding {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = json.NewDecoder(reader).Decode(output)
		if err != nil {
			err = errors.Wrap(err, "json decoding fails")
			return
		}
		defer reader.Close()
	default:
		err = json.NewDecoder(resp.Body).Decode(output)
		if err != nil {
			err = errors.Wrap(err, "json decoding fails")
			return
		}
	}
	return
}

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func makeRequest(urlStr string) (*http.Response, error) {
	os.Setenv("MARKET_DATA_KEY", "Mk5yb3RmSHZVdkt3Y2Y3cEQxSUg3dkRsSktnelI0eWI2cl80RDUxbGVoST0")
	key := os.Getenv("MARKET_DATA_KEY")
	if key == "" {
		return nil, fmt.Errorf("MARKET_DATA_KEY not found")
	}

	request, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Accept-Encoding", "gzip")
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", key))
	resp, err := httpClient.Do(request)
	return resp, err
}
