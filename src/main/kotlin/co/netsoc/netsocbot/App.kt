package co.netsoc.netsocbot
import com.jessecorbett.diskord.dsl.*
import com.jessecorbett.diskord.api.exception.DiscordBadPermissionsException
import kotlin.system.exitProcess
import co.netsoc.netsocbot.commands.init
import co.netsoc.netsocbot.commands.specialCommands
import co.netsoc.netsocbot.commands.registerDMs
import com.jessecorbett.diskord.util.*
import co.netsoc.netsocbot.utils.isDM

val BOT_TOKEN: String = System.getenv("NETSOCBOT_TOKEN")
val PREFIX = System.getenv("NETSOCBOT_PREFIX") ?: "!"

@ExperimentalUnsignedTypes
suspend fun main() {
    val environmentVariables = arrayOf("NETSOCBOT_TOKEN", "NETSOCBOT_SERVERID", "NETSOCBOT_ROLEID", "NETSOCBOT_SENDGRID_TOKEN")
    for (variable in environmentVariables) {
        if (System.getenv(variable) == null) {
            println("$variable not set\nExiting")
            exitProcess(1)
        }
    }
    val commands = init()
    bot(BOT_TOKEN) {
        commands(PREFIX) {
            for (commandString in commands.keys) {
                command(commandString) {
                    val response = commands[commandString]!!.invoke(this)
                    if(response != null) {
                        try{
                            replyAndDelete(response)
                        } catch (e: DiscordBadPermissionsException) {
                            reply(response)
                        }
                    }
                }
            }
        }
        messageCreated{
            if(it.isFromUser) {
                if(isDM(it)) {
                    registerDMs(it, this.clientStore)
                } else {
                    val function = specialCommands[it.words[0]]
                    if (function != null) {
                        function.invoke(it)
                    }
                }
            }
        }
    }
}