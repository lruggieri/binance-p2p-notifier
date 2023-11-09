package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.uber.org/atomic"

	"p2p-check/src/client/binance"
	"p2p-check/src/config"
	"p2p-check/src/event"
	"p2p-check/src/forex"
	"p2p-check/src/logger"
	"p2p-check/src/notification"
)

const (
	// MinSpamWaitTime : min mount of time to wait before re-sending a message for a specific advertiser
	MinSpamWaitTime = 5 * time.Hour

	// MinCheckTime : fx data will be fetched no more than once every MinCheckTime
	MinCheckTime = time.Minute
)

var (
	// spamFilter : avoids re-sending notifications about the same advertiser
	spamFilter = sync.Map{}
	// if paused, fx data won't be fetched
	paused = atomic.NewBool(false)

	// env variables
	configDir, slackNotificationWebhookURL, slackAppToken string
)

func populateEnvVariable() {
	configDir = os.Getenv("CONFIG_FILEPATH")
	if configDir == "" {
		panic("env CONFIG_FILEPATH not set")
	}

	slackNotificationWebhookURL = os.Getenv("SLACK_NOTIFICATION_WEBHOOK_URL")
	if slackNotificationWebhookURL == "" {
		panic("env SLACK_NOTIFICATION_WEBHOOK_URL not set")
	}

	slackAppToken = os.Getenv("SLACK_APP_TOKEN")
	if slackAppToken == "" {
		panic("env SLACK_APP_TOKEN not set")
	}
}

func main() {
	logger.InitDefault()

	populateEnvVariable()

	mainContext, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer stop()

	cfgManager, err := config.NewFileManager(mainContext, configDir)
	if err != nil {
		panic(err)
	}

	fx := forex.NewFastForexClient(os.Getenv("FASTFOREX_API_KEY"), http.DefaultClient)
	bnClient := binance.NewClient()
	notificationClient := notification.NewSlack(slackNotificationWebhookURL)
	eventClient := event.NewSlack(slackAppToken).
		WithCallback(event.Blacklist, blackListCallback(cfgManager)).
		WithCallback(event.Pause, pauseCallback()).
		WithCallback(event.Restart, restartCallback())

	errChan := make(chan error)
	resChan := make(chan float64)

	_, err = eventClient.Listen(mainContext)
	if err != nil {
		panic(err)
	}

	go startSpamFilterLoop(mainContext)

	go startErrLoop(errChan)

	go startForexLoop(mainContext, fx, errChan, resChan, cfgManager)

	go startAdvLoop(mainContext, bnClient, notificationClient, cfgManager, errChan, resChan)

	<-mainContext.Done()
	close(errChan)
	close(resChan)
}

func blackListCallback(cfgManager config.Manager) func(args string) (string, error) {
	return func(args string) (string, error) {
		generateAnswerAndSaveConfig := func(c config.Config) string {
			answer := fmt.Sprintf("Line: %s\nBank:%s",
				strings.Join(c.BlackList.Line, ","),
				strings.Join(c.BlackList.Bank, ","),
			)

			cfgManager.SaveConfig(c)

			return answer
		}

		c := cfgManager.GetConfig()

		// blacklist {name} {method}
		if args == "" {
			return generateAnswerAndSaveConfig(c), nil
		}

		argsSplit := strings.Split(args, " ")
		cleanArgs := make([]string, 0, len(argsSplit))
		for _, a := range argsSplit {
			if newArg := strings.TrimSpace(a); newArg != "" {
				cleanArgs = append(cleanArgs, newArg)
			}
		}

		if len(cleanArgs) < 2 {
			return "", errors.New("invalid arguments")
		}

		switch method := strings.ToLower(cleanArgs[1]); method {
		case "line":
			c.BlackList.Line = append(c.BlackList.Line, cleanArgs[0])
			return generateAnswerAndSaveConfig(c), nil
		case "bank":
			c.BlackList.Bank = append(c.BlackList.Bank, cleanArgs[0])
			return generateAnswerAndSaveConfig(c), nil
		default:
			return "", errors.New(fmt.Sprintf("method '%s' not supported", method))
		}
	}
}

func pauseCallback() func(args string) (string, error) {
	return func(args string) (string, error) {
		paused.Store(true)

		logger.Default.Info("PAUSE activated")

		return "paused", nil
	}
}

func restartCallback() func(args string) (string, error) {
	return func(args string) (string, error) {
		paused.Store(false)

		logger.Default.Info("PAUSE deactivated")

		return "restarted", nil
	}
}

// startSpamFilterLoop : remove elements from spamFilter older than MinSpamWaitTime
func startSpamFilterLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			spamFilter.Range(func(key, value any) bool {
				if value.(int64) < time.Now().Add(-MinSpamWaitTime).Unix() {
					spamFilter.Delete(key)
				}

				return true
			})
		}
	}
}

func startErrLoop(errChan chan error) {
	for err := range errChan {
		logger.Default.Error(err.Error())
	}
}

func startForexLoop(
	ctx context.Context,
	fc forex.Client, errChan chan error, resultChan chan float64,
	cfgManager config.Manager,
) {
	tickerTime := time.Duration(math.Max(float64(MinCheckTime), float64(fc.GetMaxRate())))
	ticker := time.NewTicker(tickerTime)

	cycle := func() {
		if paused.Load() {
			return
		}

		logger.Default.Info("fetching new rate")
		currentRate, err := fc.GetCurrentFxRate(ctx, "USD", cfgManager.GetConfig().TargetCurrency)
		if err != nil {
			errChan <- err
			return
		}
		logger.Default.WithField("rate", currentRate).Info("new rate fetched")
		resultChan <- currentRate
	}

	cycle()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cycle()
		}
	}
}

func startAdvLoop(
	ctx context.Context,
	bc *binance.Client, notificationClient notification.Client, cfgManager config.Manager,
	errChan chan error, rateChan chan float64,
) {
	var newAdvs []binance.P2PData
	var newRate float64
	var err error

	for {
		select {
		case <-ctx.Done():
			return
		case newRate = <-rateChan:
			logger.Default.Info("fetching advs")
			// TODO export as variables
			newAdvs, err = bc.ListP2PAdvs("USDT", cfgManager.GetConfig().TargetCurrency)
			if err != nil {
				errChan <- err
				continue
			}

			logger.Default.WithField("advs-size", len(newAdvs)).Info("ads fetched")

			for _, adv := range newAdvs {
				advRate, advErr := strconv.ParseFloat(adv.Adv.Price, 64)
				if advErr != nil {
					errChan <- advErr
					continue
				}

				rateSurplus := (advRate/newRate)*100 - 100

				if rateSurplus <= cfgManager.GetConfig().MaxSurplusPercentage {
					if !isAdvertiserAllowed(cfgManager, adv.Advertiser.NickName) {
						continue
					}

					// good offer
					methods := make([]string, 0, len(adv.Adv.TradeMethods))
					for _, method := range adv.Adv.TradeMethods {
						if isPayMethodAllowed(method.Identifier) {
							methods = append(methods, method.TradeMethodName)
						}
					}

					if len(methods) > 0 {
						msg := fmt.Sprintf("advertiser '%s' has a good offer."+
							"\n\tFX rate: %f"+
							"\n\tOffer rate: %f"+
							"\n\tDistance: %f"+
							"\n\tAmount: %s"+
							"\n\tMethods: %s\n",
							adv.Advertiser.NickName,
							newRate,
							advRate,
							rateSurplus,
							adv.Adv.SurplusAmount,
							strings.Join(methods, ","))
						if err = notificationClient.SendMessage(msg); err != nil {
							errChan <- advErr
							continue
						}

						spamFilter.Store(adv.Advertiser.NickName, time.Now().Unix())
					}
				}
			}
		}
	}
}

func isPayMethodAllowed(method string) bool {
	switch method {
	case "LINEPay":
		return true
	case "BANK":
		// only allowed from Monday to friday from 6AM to 14:30 PM UTC+9
		location, _ := time.LoadLocation("Asia/Tokyo") // UTC+9 corresponds to Asia/Tokyo timezone
		t := time.Now().In(location)

		hour := t.Hour()
		minute := t.Minute()

		// check if it's Monday to Friday
		if t.Weekday() >= time.Monday && t.Weekday() <= time.Friday {
			// check if it's between 6AM and 14:30 PM
			if hour > 6 && (hour < 14 || (hour == 14 && minute <= 30)) {
				return true
			}
		}
	}

	return false
}

func isAdvertiserAllowed(configManager config.Manager, advertiserNickname string) bool {
	// spam filter
	if _, ok := spamFilter.Load(advertiserNickname); ok {
		return false
	}

	// blacklist filter
	// TODO ban only for the correct payment method
	c := configManager.GetConfig()
	for _, blackListedUser := range c.BlackList.Line {
		if advertiserNickname == blackListedUser {
			return false
		}
	}
	for _, blackListedUser := range c.BlackList.Bank {
		if advertiserNickname == blackListedUser {
			return false
		}
	}

	return true
}
