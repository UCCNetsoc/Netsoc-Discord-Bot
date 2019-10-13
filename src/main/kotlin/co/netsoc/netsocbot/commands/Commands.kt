package co.netsoc.netsocbot.commands

import com.jessecorbett.diskord.dsl.Command
import com.jessecorbett.diskord.api.model.Message
import com.jessecorbett.diskord.util.*

val helpStrings = HashMap<String, String>()
val commands = HashMap<String, (Message)->String?>()

private fun register(command: String, helpString: String, action: (Message)->String?) {
    commands[command] = action
    helpStrings[command] = helpString
}

fun init(): HashMap<String, (Message)->String?> {
    register("help", "Display this message", help)
    register("ping", "Responds with a \"pong!\"", ping)


    return commands
}