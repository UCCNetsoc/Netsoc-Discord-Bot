package co.netsoc.netsocbot.utils

import co.netsoc.netsocbot.BOT_TOKEN
import com.jessecorbett.diskord.api.rest.client.*
import com.jessecorbett.diskord.api.model.*
import com.jessecorbett.diskord.util.*
import com.jessecorbett.diskord.api.rest.*
import com.sendgrid.*
import com.sendgrid.helpers.mail.objects.*
import com.sendgrid.helpers.mail.Mail
import java.security.SecureRandom

suspend fun messageUser(user: User, message: String) {
    val clientStore = ClientStore(BOT_TOKEN)
    val channel = clientStore.discord.createDM(CreateDM(user.id))
    clientStore.channels[channel.id].sendMessage(message)
}

suspend fun isDM(message: Message): Boolean {
    val clientStore = ClientStore(BOT_TOKEN)
    val channel = clientStore.discord.createDM(CreateDM(message.authorId))
    return channel.id == message.channelId
}

fun sendEmail(from: String, to: String, subject: String, content: String): Response {
    val fromAddress = Email(from)
    val toAddress = Email(to)

    val contentObj = Content("text/plain", content)
    val mail = Mail(fromAddress, subject, toAddress, contentObj)

    val sg = SendGrid(System.getenv("NETSOCBOT_SENDGRID_TOKEN") ?: "")
    val request = Request()
    request.setMethod(Method.POST)
    request.setEndpoint("mail/send")
    request.setBody(mail.build())
    return sg.api(request)
}

const val alphanum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

@ExperimentalUnsignedTypes
fun randomHash(len: Int): String {
    val random = SecureRandom()
    var bytes = ByteArray(len)
    random.nextBytes(bytes)
    var hash = ""
    for(byte in bytes) {
        val index = (byte.toUByte()).toInt()%alphanum.length
        hash += alphanum[index]
    }
    return hash
}