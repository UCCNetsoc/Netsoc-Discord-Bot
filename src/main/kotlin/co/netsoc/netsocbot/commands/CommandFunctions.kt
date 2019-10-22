package co.netsoc.netsocbot.commands

import co.netsoc.netsocbot.*
import co.netsoc.netsocbot.utils.*
import com.jessecorbett.diskord.api.model.Message
import java.net.*

fun help(message: Message): String {
    var out = "```"
    for (command in helpStrings.keys) {
        out += "${PREFIX}${command}: ${helpStrings[command]}\n"
    }
    return out + "```"
}

fun ping(message: Message): String {
    return "pong!"
}

suspend fun register(message: Message): Unit {
    val author = message.author
    if (message.guildId != null) {
        if (guilds.containsKey(author.id)) {
            guilds[author.id] += "," + message.guildId!!
        } else {
            guilds[author.id] = message.guildId!!
        }
        messageUser(author, "Please message me your UCC email address so I can verify you as a member of UCC")
    }
}