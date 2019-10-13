package Netsoc.Discord.Bot
import com.jessecorbett.diskord.dsl.*

const val BOT_TOKEN = "NDQwMTgwOTM0ODgxNTc0OTE1.XaMuIw.AKj87m-yFkBO63cNoFEAnIfXyfA"

suspend fun main() {
    bot(BOT_TOKEN) {
        commands {
            command("ping") {
                reply("pong")
            }
        }
    }
}
