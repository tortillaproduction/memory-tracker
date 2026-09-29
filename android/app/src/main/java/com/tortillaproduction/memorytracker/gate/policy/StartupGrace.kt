package com.tortillaproduction.memorytracker.gate.policy

/**
 * 端末の起動直後・アプリ更新の直後の一定時間はゲートを出さない。
 * 起動処理中やアップデート直後の不安定な状態で画面を奪わないため。
 */
object StartupGrace {
    const val DURATION_MILLIS = 60_000L

    /**
     * @param sinceBootMillis 端末起動からの経過時間(SystemClock.elapsedRealtime)
     * @param sinceUpdateMillis アプリのインストール・更新からの経過時間
     */
    fun isInGrace(sinceBootMillis: Long, sinceUpdateMillis: Long): Boolean =
        sinceBootMillis < DURATION_MILLIS || sinceUpdateMillis in 0 until DURATION_MILLIS
}
