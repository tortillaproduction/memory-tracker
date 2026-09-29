package com.tortillaproduction.memorytracker.gate

import android.content.Context
import com.tortillaproduction.memorytracker.gate.policy.RateLimiter
import java.time.Instant
import java.time.ZoneId

/**
 * ゲートの安全装置の状態を端末に保存する。サービスが再起動しても発動制限や
 * 「今日は済み」が消えないよう、同期的に読めるSharedPreferencesに置く。
 */
class GuardStore(context: Context) {
    private val prefs = context.getSharedPreferences("gate_guard", Context.MODE_PRIVATE)

    /** 今日(日本時間)。サーバーの判定と同じ日付の区切りにする。 */
    fun today(now: Long = System.currentTimeMillis()): String =
        Instant.ofEpochMilli(now).atZone(ZONE).toLocalDate().toString()

    /** 今日すでにチェックインしたか、脱出口で解除したか。 */
    fun isFinishedToday(now: Long = System.currentTimeMillis()): Boolean {
        val today = today(now)
        return prefs.getString(KEY_DONE_ON, null) == today || prefs.getString(KEY_DISMISSED_ON, null) == today
    }

    fun markDoneToday() = prefs.edit().putString(KEY_DONE_ON, today()).apply()

    /** 防御策2: 脱出口を使った記録(その日は再発動しない)。 */
    fun markDismissedToday() = prefs.edit().putString(KEY_DISMISSED_ON, today()).apply()

    /** サーバーへ未送信の解除記録。日付が変わったら送らずに捨てる(サーバーの記録も日単位のため)。 */
    fun markDismissalPending() = prefs.edit().putString(KEY_PENDING_DISMISSAL, today()).apply()

    fun hasPendingDismissalForToday(): Boolean = prefs.getString(KEY_PENDING_DISMISSAL, null) == today()

    fun clearPendingDismissal() = prefs.edit().remove(KEY_PENDING_DISMISSAL).apply()

    fun recordGateShown(now: Long = System.currentTimeMillis()) = prefs.edit().putLong(KEY_LAST_SHOWN, now).apply()

    fun lastGateShownAt(): Long? = prefs.getLong(KEY_LAST_SHOWN, 0).takeIf { it > 0 }

    fun loadRateLimiter(): RateLimiter {
        val limiter = RateLimiter()
        val lastFireByApp = prefs.getString(KEY_LAST_FIRE, "").orEmpty()
            .split(';').filter { it.contains('=') }
            .associate { it.substringBefore('=') to (it.substringAfter('=').toLongOrNull() ?: 0L) }
        val recent = prefs.getString(KEY_RECENT, "").orEmpty()
            .split(',').mapNotNull { it.toLongOrNull() }
        limiter.restore(RateLimiter.Snapshot(lastFireByApp, recent, prefs.getLong(KEY_SUSPENDED_UNTIL, 0)))
        return limiter
    }

    fun saveRateLimiter(limiter: RateLimiter) {
        val s = limiter.snapshot()
        prefs.edit()
            .putString(KEY_LAST_FIRE, s.lastFireByApp.entries.joinToString(";") { "${it.key}=${it.value}" })
            .putString(KEY_RECENT, s.recentFires.joinToString(","))
            .putLong(KEY_SUSPENDED_UNTIL, s.suspendedUntil)
            .apply()
    }

    fun suspendedUntil(): Long = prefs.getLong(KEY_SUSPENDED_UNTIL, 0)

    companion object {
        val ZONE: ZoneId = ZoneId.of("Asia/Tokyo")
        private const val KEY_DONE_ON = "done_on"
        private const val KEY_DISMISSED_ON = "dismissed_on"
        private const val KEY_LAST_SHOWN = "last_shown_at"
        private const val KEY_PENDING_DISMISSAL = "pending_dismissal_on"
        private const val KEY_LAST_FIRE = "rate_last_fire"
        private const val KEY_RECENT = "rate_recent"
        private const val KEY_SUSPENDED_UNTIL = "rate_suspended_until"
    }
}
