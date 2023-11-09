package forex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"p2p-check/src/client/httpclient"
	"p2p-check/src/logger"
)

const (
	APIURL           = "https://api.fastforex.io"
	FetchOneEndpoint = "fetch-one"
	FetchAllEndpoint = "fetch-all"
)

type clientFetchOneResponse struct {
	Error   string             `json:"error"`
	Base    string             `json:"base"`
	Result  map[string]float64 `json:"result"`
	Updated string             `json:"updated"`
	Ms      int                `json:"ms"`
}

type clientFetchAllResponse struct {
	Error   string             `json:"error"`
	Base    string             `json:"base"`
	Results map[string]float64 `json:"results"`
	Updated string             `json:"updated"`
	Ms      int                `json:"ms"`
}

type FastForexClient struct {
	APIKey     string
	HTTPClient httpclient.Client
}

func (i *FastForexClient) GetMaxRate() time.Duration {
	return time.Minute / 2
}

func (i *FastForexClient) GetCurrentFxRate(ctx context.Context, fromCurrency, toCurrency string) (float64, error) {
	url := fmt.Sprintf("%s/%s?from=%s&to=%s&api_key=%s",
		APIURL,
		FetchOneEndpoint,
		fromCurrency,
		toCurrency,
		i.APIKey,
	)

	var response clientFetchOneResponse
	if err := i.fetch(ctx, url, &response); err != nil {
		return 0, errors.Wrap(err, "cannot fetch rate")
	}

	rate, ok := response.Result[strings.ToUpper(toCurrency)]
	if !ok {
		logger.Default.WithField("response", response).Error("result is incorrect")

		return 0, errors.New("result is incorrect")
	}

	return rate, nil
}

func (i *FastForexClient) fetch(ctx context.Context, url string, resp interface{}) error {
	httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	httpReq.Header.Add("accept", "application/json")

	httpResp, err := i.HTTPClient.Do(httpReq)
	if err != nil {
		return err
	}

	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {
		logger.Default.WithField("status-code", httpResp.StatusCode).Error("status code != 200")

		return errors.New("status code != 200")
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return errors.Wrap(err, "cannot read response body")
	}

	if err = json.Unmarshal(body, resp); err != nil {
		logger.Default.WithError(err).WithField("response", string(body)).Error("invalid service response")

		return errors.Wrap(err, "invalid service response")
	}

	return nil
}

func NewFastForexClient(apiKey string, httpClient httpclient.Client) Client {
	return &FastForexClient{
		APIKey:     apiKey,
		HTTPClient: httpClient,
	}
}
