package forex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"p2p-check/src/logger"
)

const (
	AlphavantageHost   = "https://www.alphavantage.co"
	AlphavantageAPIKEY = "1R4TFWPP0H5DT2CC"
)

type Alphavantage struct{}

func (av *Alphavantage) GetCurrentFxRate(_ context.Context, fromCurrency, toCurrency string) (float64, error) {
	resp, err := http.Get(
		fmt.Sprintf("%s/query?function=CURRENCY_EXCHANGE_RATE&from_currency=%s&to_currency=%s&apikey=%s",
			AlphavantageHost, fromCurrency, toCurrency, AlphavantageAPIKEY))
	if err != nil {
		return 0, err
	}

	var respBytes []byte
	if resp.Body != nil {
		respBytes, _ = io.ReadAll(resp.Body)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Default.
			WithField("resp", string(respBytes)).
			WithField("status-code", resp.StatusCode).
			Error("status code != 200")
		return 0, errors.New("status code != 200")
	}

	type r struct {
		Rate struct {
			From string `json:"1. From_Currency Code"`
			To   string `json:"3. To_Currency Code"`
			Rate string `json:"5. Exchange Rate"`
		} `json:"Realtime Currency Exchange Rate"`
		Error string `json:"Error Message"`
	}

	var rateResp r
	_ = json.Unmarshal(respBytes, &rateResp)

	if rateResp.Error != "" {
		logger.Default.
			WithField("resp", string(respBytes)).
			Error(rateResp.Error)
		return 0, errors.Wrap(errors.New(rateResp.Error), "rate APi returned error")
	}

	rateFloat, err := strconv.ParseFloat(rateResp.Rate.Rate, 64)
	if err != nil {
		logger.Default.
			WithField("resp", string(respBytes)).
			Error(err.Error())
		return 0, errors.Wrap(err, "cannot convert rate to float")
	}

	return rateFloat, nil
}

func (av *Alphavantage) GetMaxRate() time.Duration {
	return time.Minute / 5
}

func NewAlphavantage() Client {
	return &Alphavantage{}
}
