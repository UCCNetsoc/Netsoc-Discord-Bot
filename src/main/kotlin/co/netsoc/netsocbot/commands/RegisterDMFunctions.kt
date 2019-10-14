package co.netsoc.netsocbot.commands

import com.jessecorbett.diskord.api.model.Message
import com.jessecorbett.diskord.util.*
import co.netsoc.netsocbot.utils.sendEmail
import co.netsoc.netsocbot.utils.randomHash

val hashes = HashMap<String, String>()

@ExperimentalUnsignedTypes
suspend fun registerDMs(message: Message, clientStore: ClientStore) {
    val word = message.words[0]
    if (word.endsWith("@umail.ucc.ie")) {
        val randomHash = randomHash(64)
        val response = sendEmail("server.registration@netsoc.co", word, "Netsoc Discord Verification", "Please message the following token to the Netsoc Bot to gain access to the Discord Server:\n\n" + randomHash + "\n\nIf you did not request access to the Netsoc Discord Server, ignore this message.")
        if (response.statusCode == 200 || response.statusCode == 202) {
            hashes[message.authorId] = randomHash
        }
        clientStore.channels[message.channelId].sendMessage("Please reply with the token that has been emailed to you")
    } else if(hashes.containsKey(message.authorId)) {
        if (word == hashes[message.authorId]) {
            clientStore.channels[message.channelId].sendMessage("Thank you. You have been register for the Netsoc Discord Server")
            clientStore.guilds[System.getenv("NETSOCBOT_SERVERID")].addMemberRole(message.authorId, System.getenv("NETSOCBOT_ROLEID"))
        }
    } else {
        clientStore.channels[message.channelId].sendMessage("Please use a valid UCC email address")
    }
}