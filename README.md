# Netsoc Discord Bot ![Build Status](https://ci.netsoc.co/api/badges/UCCNetworkingSociety/Netsoc-Discord-Bot/status.svg?branch=master)

The Netsoc Discord bot is intended to help make the lives of Netsoc SysAdmins easier. This is done by:
* Integration with [Netsoc Admin](https://admin.netsoc.co)
* Relaying server alerts to discord
* Various commands which members of the Netsoc Discord Server can use
* And more upcoming cool stuff B)

## Current Commands:

* ping - will reply "Pong!"
* alias - Sets a shortcut command. Usage: `!alias keyword url_link_to_resource`
* help - If followed by a command name, it shows the details of the command
* info - Displays some info about NetsocBot
* config - Displays the config for NetsocBot
* minecraft - Displays the number of people online in the Netsoc Minecraft Server
* inspire - Displays an inspirational quote 

## Config file

```json
{
	"token": "AAABBBccccc1111",
	"prefix": "!",
	"helpChannelId": "123456789",
	"alertsChannelId": "123456789",
	"sysAdminTag": "<@&123456789>",
	"botHostName": "0.0.0.0:4201",
	"guildID": "123456789",
	"permissions": {
		"ping" : [
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
| `"helpChannelID"` | The channel ID to which help messages from Netsoc Admin will be sent | `"alertsChannelId"` | The channel ID to which firing alerts from the server will be relayed |
| `"sysAdminTag"` | The ID which will be @'d for help requests and alerting |
| `"botHostName"` | The host name and port which the bot will listen on for help and alert webhooks |
| `"guildID"` | The ID of the guild (i.e. server) of which this bot is a member | 
| `"permissions"` | Indicates the roles/users which have permission to run perticular commands. If none are specified for command, anyone can run the command |

