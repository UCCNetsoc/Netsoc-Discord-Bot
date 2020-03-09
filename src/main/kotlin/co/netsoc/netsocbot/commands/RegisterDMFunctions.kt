package co.netsoc.netsocbot.commands

import com.jessecorbett.diskord.api.model.Message
import com.jessecorbett.diskord.util.*
import com.jessecorbett.diskord.api.exception.DiscordException
import co.netsoc.netsocbot.utils.sendEmail
import co.netsoc.netsocbot.utils.randomHash
import co.netsoc.netsocbot.ROLEIDS

val guilds = HashMap<String, String>()
val hashes = HashMap<String, String>()

@ExperimentalUnsignedTypes
suspend fun registerDMs(message: Message, clientStore: ClientStore) {
    val word = message.words[0]
    if (guilds[message.authorId] == null) {
        return
    }

    // If they haven't been sent a hash value - get email and send hash
    if (hashes[message.authorId] == null) {
        if (!word.endsWith("@umail.ucc.ie")) {
            clientStore.channels[message.channelId].sendMessage("Please use a valid UCC email address")
            return
        }

        val randomHash = randomHash(64)
        val response = sendEmail("server.registration@netsoc.co", word, "Netsoc Discord Verification", "Please message the following token to the Netsoc Bot to gain access to the Discord Server:\n\n" + randomHash + "\n\nIf you did not request access to the Netsoc Discord Server, ignore this message.")
        if (response.statusCode == 200 || response.statusCode == 202) {
            hashes[message.authorId] = randomHash
            clientStore.channels[message.channelId].sendMessage("Please reply with the token that has been emailed to you")
        } else {
            clientStore.channels[message.channelId].sendMessage("Failed to send email. Please try again later")
        }
        return
    }

    // Response is not the same as given hash
    if (word != hashes[message.authorId]) {
        clientStore.channels[message.channelId].sendMessage("Incorrect token. Please try again")
        return
    }

    var success = false
    for (guildid in guilds[message.authorId]!!.split(",")) {
        for (roleid in ROLEIDS) {
            try {
                clientStore.guilds[guildid].addMemberRole(message.authorId, roleid)
            } catch(e: DiscordException) {}
        }
    }
    if (!success) {
        clientStore.channels[message.channelId].sendMessage("Failed to register for the server. Please contact the owners of the server")
        return
    }
    clientStore.channels[message.channelId].sendMessage("Thank you. You have been registered for the Netsoc Discord Server")
    hashes.remove(message.authorId)
}