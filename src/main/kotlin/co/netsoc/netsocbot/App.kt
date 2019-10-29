package co.netsoc.netsocbot
import com.jessecorbett.diskord.dsl.*
import com.jessecorbett.diskord.api.exception.*
import kotlin.system.exitProcess
import co.netsoc.netsocbot.commands.init
import co.netsoc.netsocbot.commands.specialCommands
import co.netsoc.netsocbot.commands.registerDMs
import com.jessecorbett.diskord.util.*
import co.netsoc.netsocbot.utils.isDM
import java.net.UnknownHostException

val BOT_TOKEN: String = System.getenv("NETSOCBOT_TOKEN")
val PREFIX = System.getenv("NETSOCBOT_PREFIX") ?: "!"
val ROLEIDS = (System.getenv("NETSOCBOT_ROLEIDS") ?: "").split(",")

suspend fun setup() {
    val environmentVariables = arrayOf("NETSOCBOT_TOKEN", "NETSOCBOT_ROLEIDS", "NETSOCBOT_SENDGRID_TOKEN")
    for (variable in environmentVariables) {
        if (System.getenv(variable) == null) {
            println("$variable not set\nExiting")
            exitProcess(1)
        }
    }

}

@ExperimentalUnsignedTypes
suspend fun main() {
    setup()
    val commands = init()
    try {
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
    } catch (e: DiscordUnauthorizedException) {
        println(e)
        exitProcess(1)
    } catch (e: UnknownHostException) {
        println(e)
        exitProcess(1)
    }
}
