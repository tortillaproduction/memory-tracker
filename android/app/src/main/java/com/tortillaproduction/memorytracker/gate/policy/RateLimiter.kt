package com.tortillaproduction.memorytracker.gate.policy

/**
 * 防御策4: 発動制限。他の防御策とは独立に、暴走・無限ループを止める。
 *  - 同じアプリでの再発動は [perAppInterval] 以上空ける
 *  - [burstWindow] の間に [burstLimit] 回を超えて発動しようとしたら、[suspendDuration] の間すべて止める
 *
 * 時刻はミリ秒(System.currentTimeMillis)。状態は [snapshot] / [restore] で永続化できる。
 */
class RateLimiter(
    private val perAppInterval: Long = 5 * MINUTE,
    private val burstWindow: Long = 10 * MINUTE,
    private val burstLimit: Int = 5,
    private val suspendDuration: Long = 60 * MINUTE,
) {
    data class Snapshot(
        val lastFireByApp: Map<String, Long>,
        val recentFires: List<Long>,
        val suspendedUntil: Long,
    )

    private val lastFireByApp = mutableMapOf<String, Long>()
    private val recentFires = ArrayDeque<Long>()
    var suspendedUntil: Long = 0
        private set

    /**
     * 発動の見込みがあるかを、状態を変えずに確かめる(停止中か、同じアプリの最小間隔内ならfalse)。
     * サーバーに候補を問い合わせる前の足切りに使い、実際に出すときに [tryAcquire] で記録する。
     */
    fun canAcquire(packageName: String, now: Long): Boolean {
        if (now < suspendedUntil) return false
        val last = lastFireByApp[packageName]
        return last == null || now - last >= perAppInterval
    }

    /** 発動してよければ記録してtrueを返す。 */
    fun tryAcquire(packageName: String, now: Long): Boolean {
        if (now < suspendedUntil) return false

        val last = lastFireByApp[packageName]
        if (last != null && now - last < perAppInterval) return false

        while (recentFires.isNotEmpty() && now - recentFires.first() >= burstWindow) {
            recentFires.removeFirst()
        }
        if (recentFires.size >= burstLimit) {
            suspendedUntil = now + suspendDuration
            recentFires.clear()
            return false
        }

        recentFires.addLast(now)
        lastFireByApp[packageName] = now
        return true
    }

    fun isSuspended(now: Long): Boolean = now < suspendedUntil

    fun snapshot(): Snapshot = Snapshot(lastFireByApp.toMap(), recentFires.toList(), suspendedUntil)

    fun restore(s: Snapshot) {
        lastFireByApp.clear()
        lastFireByApp.putAll(s.lastFireByApp)
        recentFires.clear()
        recentFires.addAll(s.recentFires.sorted())
        suspendedUntil = s.suspendedUntil
    }

    companion object {
        const val MINUTE = 60_000L
    }
}
