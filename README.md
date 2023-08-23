## Binance-P2P-Notifier
A small project with the purpose of being notified when there are P2P offers to trade USDT with a specific currency
{targetCurrency} that are below a threshold percentage wrt the current forex price for that currency (USD-{targetCurrency}).
Provides enough abstraction to implement your FX provider, notification and event method.

### Config File
Create a JSON config file with the following structure:
```json
{
  "blackList": {
    "line": [],
    "bank": []
  },
  "maxSurplusPercentage": 1,
  "targetCurrency": "JPY"
}
```

### Environment variables
* CONFIG_FILEPATH: absolute path to the JSON configuration file
* SLACK_NOTIFICATION_WEBHOOK_URL: Slack hook URL (https://api.slack.com/messaging/webhooks).
  <br/>Format 'https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX'
* SLACK_APP_TOKEN= Slack app-level token (https://api.slack.com/authentication/token-types).
  <br/>Format 'xapp-*'
