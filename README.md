# Netsoc Discord Bot [![Build Status](https://travis-ci.org/UCCNetworkingSociety/Netsoc-Discord-Bot.svg)](https://travis-ci.org/UCCNetworkingSociety/Netsoc-Discord-Bot)

## Netsoc Admin Integration

Help messages which are sent from Netsoc Admin will appear in the
specified help channel.

## Current Commands:

* !ping - will reply "Pong!"
* !alias - Sets a shortcut command. Usage: `!alias keyword url_link_to_resource`
* !help - If followed by a command name, it shows the details of the command
* !top - Prints the output of `top -b -n 1`
* !sensors - Displays temperature of the server
* !info - Displays some info about NetsocBot
* !config - Displays the config for NetsocBot

## Config file

```json
{
    "token": "AAABBBccccc1111",
	"prefix": "!",
	"helpChannelId": "357195183999287296", 
	"sysAdminTag": "<@&318907623476822016>",
	"botHostName": "0.0.0.0:4201",
	"guildID": "291573897730588684",
	"permissions": {
		"set" : [
			"SysAdmin",
			"HLM"
		]
	}
}
```

| Config Value      | Purpose                                  |
| ----------------- | ---------------------------------------- |
| `"token"`         | The authentication token used by the Discord bot |
| `"prefix"`        | The string that prefixes all commands    |
| `"helpChannelID"` | The channel ID to which help messages from Netsoc Admin will be sent |
