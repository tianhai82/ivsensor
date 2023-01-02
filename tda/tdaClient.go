package tda

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/tianhai82/ivsensor/model"
)

const FrequencyDaily = "daily"
const FrequencyWeekly = "weekly"

const OptionContractPUT = "PUT"
const OptionContractCALL = "CALL"
const OptionContractALL = "ALL"

const OptionRangeAll = "ALL"
const OptionRangeITM = "ITM"
const OptionRangeOTM = "OTM"
const OptionRangeNTM = "NTM"

const baseURL = "https://stock-timing.appspot.com/rpc/tiger"

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func RetrieveHistory(symbol string, freqencyType string, period int) ([]model.Candle, error) {
	url := fmt.Sprintf("%s/pricehistory/%s/%s/%d", baseURL, symbol, freqencyType, period)
	var candles []model.Candle
	err := makeGetRequest(url, &candles)
	return candles, err
}

func RetrieveOptions(symbol, contractType, optionRange, fromDate, toDate string) (model.Chains, error) {
	url := fmt.Sprintf("%s/optionchain/%s/%s/%s/%s/%s", baseURL, symbol, contractType, optionRange, fromDate, toDate)
	var chain model.Chains
	err := makeGetRequest(url, &chain)
	return chain, err
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

func makeRequest(urlStr string) (*http.Response, error) {
	key := os.Getenv("TDA_KEY")
	if key == "" {
		return nil, fmt.Errorf("TDA_KEY not found")
	}

	request, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Accept-Encoding", "gzip")

	ts := time.Now().UTC().Unix()

	payload := strconv.FormatInt(ts, 10)
	payloadEncoded := base64.RawURLEncoding.EncodeToString([]byte(payload))

	signature := hmac256hash([]byte(payload), []byte(key))
	signatureEncoded := base64.RawURLEncoding.EncodeToString([]byte(signature))

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s.%s", payloadEncoded, signatureEncoded))
	resp, err := httpClient.Do(request)
	return resp, err
}

func hmac256hash(msg []byte, key []byte) []byte {
	sig := hmac.New(sha256.New, key)
	sig.Write([]byte(msg))
	return []byte(hex.EncodeToString(sig.Sum(nil)))
}
