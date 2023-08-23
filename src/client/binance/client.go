package binance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"p2p-check/src/logger"
)

const (
	P2PHost = "https://p2p.binance.com"
)

type P2PData struct {
	Adv struct {
		AdvNo                 string `json:"advNo"`
		Classify              string `json:"classify"`
		TradeType             string `json:"tradeType"`
		Asset                 string `json:"asset"`
		FiatUnit              string `json:"fiatUnit"`
		AdvStatus             any    `json:"advStatus"`
		PriceType             any    `json:"priceType"`
		PriceFloatingRatio    any    `json:"priceFloatingRatio"`
		RateFloatingRatio     any    `json:"rateFloatingRatio"`
		CurrencyRate          any    `json:"currencyRate"`
		Price                 string `json:"price"`
		InitAmount            any    `json:"initAmount"`
		SurplusAmount         string `json:"surplusAmount"`
		AmountAfterEditing    any    `json:"amountAfterEditing"`
		MaxSingleTransAmount  string `json:"maxSingleTransAmount"`
		MinSingleTransAmount  string `json:"minSingleTransAmount"`
		BuyerKycLimit         any    `json:"buyerKycLimit"`
		BuyerRegDaysLimit     any    `json:"buyerRegDaysLimit"`
		BuyerBtcPositionLimit any    `json:"buyerBtcPositionLimit"`
		Remarks               any    `json:"remarks"`
		AutoReplyMsg          string `json:"autoReplyMsg"`
		PayTimeLimit          int    `json:"payTimeLimit"`
		TradeMethods          []struct {
			PayID                any    `json:"payId"`
			PayMethodID          string `json:"payMethodId"`
			PayType              any    `json:"payType"`
			PayAccount           any    `json:"payAccount"`
			PayBank              any    `json:"payBank"`
			PaySubBank           any    `json:"paySubBank"`
			Identifier           string `json:"identifier"`
			IconURLColor         any    `json:"iconUrlColor"`
			TradeMethodName      string `json:"tradeMethodName"`
			TradeMethodShortName string `json:"tradeMethodShortName"`
			TradeMethodBgColor   string `json:"tradeMethodBgColor"`
		} `json:"tradeMethods"`
		UserTradeCountFilterTime        any    `json:"userTradeCountFilterTime"`
		UserBuyTradeCountMin            any    `json:"userBuyTradeCountMin"`
		UserBuyTradeCountMax            any    `json:"userBuyTradeCountMax"`
		UserSellTradeCountMin           any    `json:"userSellTradeCountMin"`
		UserSellTradeCountMax           any    `json:"userSellTradeCountMax"`
		UserAllTradeCountMin            any    `json:"userAllTradeCountMin"`
		UserAllTradeCountMax            any    `json:"userAllTradeCountMax"`
		UserTradeCompleteRateFilterTime any    `json:"userTradeCompleteRateFilterTime"`
		UserTradeCompleteCountMin       any    `json:"userTradeCompleteCountMin"`
		UserTradeCompleteRateMin        any    `json:"userTradeCompleteRateMin"`
		UserTradeVolumeFilterTime       any    `json:"userTradeVolumeFilterTime"`
		UserTradeType                   any    `json:"userTradeType"`
		UserTradeVolumeMin              any    `json:"userTradeVolumeMin"`
		UserTradeVolumeMax              any    `json:"userTradeVolumeMax"`
		UserTradeVolumeAsset            any    `json:"userTradeVolumeAsset"`
		CreateTime                      any    `json:"createTime"`
		AdvUpdateTime                   any    `json:"advUpdateTime"`
		FiatVo                          any    `json:"fiatVo"`
		AssetVo                         any    `json:"assetVo"`
		AdvVisibleRet                   any    `json:"advVisibleRet"`
		AssetLogo                       any    `json:"assetLogo"`
		AssetScale                      int    `json:"assetScale"`
		FiatScale                       int    `json:"fiatScale"`
		PriceScale                      int    `json:"priceScale"`
		FiatSymbol                      string `json:"fiatSymbol"`
		IsTradable                      bool   `json:"isTradable"`
		DynamicMaxSingleTransAmount     string `json:"dynamicMaxSingleTransAmount"`
		MinSingleTransQuantity          string `json:"minSingleTransQuantity"`
		MaxSingleTransQuantity          string `json:"maxSingleTransQuantity"`
		DynamicMaxSingleTransQuantity   string `json:"dynamicMaxSingleTransQuantity"`
		TradableQuantity                string `json:"tradableQuantity"`
		CommissionRate                  string `json:"commissionRate"`
		TradeMethodCommissionRates      []any  `json:"tradeMethodCommissionRates"`
		LaunchCountry                   any    `json:"launchCountry"`
		AbnormalStatusList              any    `json:"abnormalStatusList"`
		CloseReason                     any    `json:"closeReason"`
		StoreInformation                any    `json:"storeInformation"`
	} `json:"adv"`
	Advertiser struct {
		UserNo             string  `json:"userNo"`
		RealName           string  `json:"realName"`
		NickName           string  `json:"nickName"`
		Margin             any     `json:"margin"`
		MarginUnit         any     `json:"marginUnit"`
		OrderCount         any     `json:"orderCount"`
		MonthOrderCount    int     `json:"monthOrderCount"`
		MonthFinishRate    float64 `json:"monthFinishRate"`
		PositiveRate       float64 `json:"positiveRate"`
		AdvConfirmTime     any     `json:"advConfirmTime"`
		Email              any     `json:"email"`
		RegistrationTime   any     `json:"registrationTime"`
		Mobile             any     `json:"mobile"`
		UserType           string  `json:"userType"`
		TagIconUrls        []any   `json:"tagIconUrls"`
		UserGrade          int     `json:"userGrade"`
		UserIdentity       string  `json:"userIdentity"`
		ProMerchant        any     `json:"proMerchant"`
		IsBlocked          any     `json:"isBlocked"`
		ActiveTimeInSecond int     `json:"activeTimeInSecond"`
	} `json:"advertiser"`
}

type P2PAdvResponse struct {
	Code          string    `json:"code"`
	Message       any       `json:"message"`
	MessageDetail any       `json:"messageDetail"`
	Data          []P2PData `json:"data"`
	Total         int       `json:"total"`
	Success       bool      `json:"success"`
}

type P2PAdvRequest struct {
	ProMerchantAds bool   `json:"proMerchantAds"`
	Page           int    `json:"page,omitempty"`
	Rows           int    `json:"rows,omitempty"`
	PayTypes       []any  `json:"payTypes,omitempty"`
	Countries      []any  `json:"countries,omitempty"`
	PublisherType  any    `json:"publisherType,omitempty"`
	Asset          string `json:"asset,omitempty"`
	Fiat           string `json:"fiat,omitempty"`
	TradeType      string `json:"tradeType,omitempty"`
}

type Client struct{}

func (c *Client) ListP2PAdvs(asset, fiat string) ([]P2PData, error) {
	req := P2PAdvRequest{
		Page:      1,
		Rows:      20,
		Asset:     asset,
		Fiat:      fiat,
		TradeType: "BUY",
	}

	reqBytes, _ := json.Marshal(req)
	resp, err := http.Post(
		fmt.Sprintf("%s/bapi/c2c/v2/friendly/c2c/adv/search", P2PHost),
		"application/json",
		io.NopCloser(bytes.NewReader(reqBytes)),
	)

	if err != nil {
		logger.Default.WithError(err).Error("ListP2PAdvs error")
		return nil, errors.Wrap(err, "ListP2PAdvs error")
	}

	var respBytes []byte
	if resp.Body != nil {
		respBytes, _ = io.ReadAll(resp.Body)
	}

	if resp.StatusCode != 200 {
		logger.Default.WithField("resp", string(respBytes)).Error("ListP2PAdvs status code != 200")
		return nil, errors.New("ListP2PAdvs status code != 200")
	}

	var r P2PAdvResponse
	if err = json.Unmarshal(respBytes, &r); err != nil {
		logger.Default.WithField("resp", string(respBytes)).Error("ListP2PAdvs invalid resp")
		return nil, errors.New("ListP2PAdvs invalid resp")
	}

	if !r.Success {
		logger.Default.WithField("resp", string(respBytes)).Error("ListP2PAdvs no success")
		return nil, errors.New("ListP2PAdvs no success")
	}

	return r.Data, nil
}

func NewClient() *Client {
	return &Client{}
}
