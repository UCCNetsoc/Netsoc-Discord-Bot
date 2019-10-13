package co.netsoc.netsocbot
import com.jessecorbett.diskord.dsl.*
import com.jessecorbett.diskord.api.exception.DiscordBadPermissionsException
import kotlin.system.exitProcess
import co.netsoc.netsocbot.commands.init

val BOT_TOKEN: String = System.getenv("NETSOCBOT_TOKEN") ?: exitProcess(1)
val PREFIX = System.getenv("NETSOCBOT_PREFIX") ?: "!"

suspend fun main() {
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
    }
}
