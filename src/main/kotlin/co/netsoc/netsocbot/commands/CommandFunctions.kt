package co.netsoc.netsocbot.commands

import com.jessecorbett.diskord.api.model.Message

const val minecraftAPIURL = "http://minecraft.netsoc.co/standalone/dynmap_NetsocCraft.json"

val help: (message: Message) -> String = {
    var out = "```"
    for (command in helpStrings.keys) {
        out += "${command}: ${helpStrings[command]}\n"
    }
    out + "```"
}

val ping: (message: Message) -> String = {
    "pong!"
}