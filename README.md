# Netsoc Discord Bot [![Build Status](https://travis-ci.org/UCCNetworkingSociety/Netsoc-Discord-Bot.svg)](https://travis-ci.org/UCCNetworkingSociety/Netsoc-Discord-Bot)

## Netsoc Admin Integration

Help messages which are sent from Netsoc Admin will appear in the
specified help channel.

## Current Commands:

* !ping - will reply "Pong!"

## Config file

```json
{
    "token": "hgjkhgkjh.oihojhkhk.iughjhgbjhvjv",
    "prefix": "!",
    "helpChannelId": "876868979834798273" 
}
```

| Config Value      | Purpose                                  |
| ----------------- | ---------------------------------------- |
| `"token"`         | The authentication token used by the Discord bot |
| `"prefix"`        | The string that prefixes all commands    |
| `"helpChannelID"` | The channel ID to which help messages from Netsoc Admin will be sent |
