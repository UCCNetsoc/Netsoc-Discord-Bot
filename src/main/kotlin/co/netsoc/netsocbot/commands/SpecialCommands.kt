package co.netsoc.netsocbot.commands

import com.jessecorbett.diskord.api.model.Message
import co.netsoc.netsocbot.PREFIX

val specialCommands = HashMap<String, suspend (Message) -> Unit>()

internal fun specialRegister(command: String, helpString: String, action: suspend (Message)->Unit) {
    val prefixedCommand = "$PREFIX" + command
    specialCommands[prefixedCommand] = action
    helpStrings[prefixedCommand] = helpString
}