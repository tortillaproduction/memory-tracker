package com.tortillaproduction.memorytracker.gate.api

import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

/** バックエンドへの接続情報。tokenは秘匿情報なので、ログや画面に出さない。 */
data class Credentials(val baseUrl: String, val token: String) {
    override fun toString(): String = "Credentials(baseUrl=$baseUrl, token=***)"
}

/**
 * WebのゲートセットアップでQRコードに埋め込まれる文字列(セットアップコード)を解釈する。
 * 形式: {"v":1,"baseUrl":"https://...","token":"..."}
 */
object SetupCode {
    @Serializable
    private data class Payload(val v: Int, val baseUrl: String, val token: String)

    private val json = Json { ignoreUnknownKeys = true }

    /** 解釈できなければnullを返す。 */
    fun parse(raw: String): Credentials? {
        val payload = runCatching { json.decodeFromString<Payload>(raw.trim()) }.getOrNull() ?: return null
        if (payload.v != 1 || payload.token.isBlank()) return null
        val baseUrl = payload.baseUrl.trim().trimEnd('/')
        if (!baseUrl.startsWith("https://") && !baseUrl.startsWith("http://")) return null
        return Credentials(baseUrl = baseUrl, token = payload.token.trim())
    }
}
