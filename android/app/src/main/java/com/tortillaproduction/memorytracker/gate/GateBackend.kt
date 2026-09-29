package com.tortillaproduction.memorytracker.gate

import android.content.Context
import com.tortillaproduction.memorytracker.gate.api.Credentials
import com.tortillaproduction.memorytracker.gate.api.GateClient
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch

/**
 * プロセス全体で共有するバックエンド接続。画面(Activity)が閉じても送信を続けられるよう、
 * 独自のコルーチンスコープを持つ。
 */
object GateBackend {
    val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    @Volatile
    private var cached: Pair<Credentials, GateClient>? = null

    /** 接続情報が未設定ならnull(ゲートは出さずに素通しする)。 */
    suspend fun client(context: Context): GateClient? {
        val creds = CredentialsStore(context).current() ?: return null
        cached?.let { (c, client) -> if (c == creds) return client }
        return GateClient(creds).also { cached = creds to it }
    }

    /**
     * 防御策2: 脱出口で解除したことをサーバーに送る。端末の記録は呼び出し側で先に済ませておく。
     * 送れなかった場合は保留にして、次にAPIを呼ぶときに再送する(その日のうちだけ)。
     */
    fun sendDismissal(context: Context) {
        val app = context.applicationContext
        val store = GuardStore(app)
        store.markDismissalPending()
        scope.launch { flushPendingDismissal(app) }
    }

    suspend fun flushPendingDismissal(context: Context) {
        val store = GuardStore(context)
        if (!store.hasPendingDismissalForToday()) {
            store.clearPendingDismissal()
            return
        }
        val client = client(context) ?: return
        if (client.dismiss()) store.clearPendingDismissal()
    }
}
