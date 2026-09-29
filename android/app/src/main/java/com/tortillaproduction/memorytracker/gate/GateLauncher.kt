package com.tortillaproduction.memorytracker.gate

import android.accessibilityservice.AccessibilityService
import android.content.Context
import android.content.Intent
import android.os.Handler
import android.os.HandlerThread
import android.os.Looper
import com.tortillaproduction.memorytracker.gate.ui.GateActivity
import java.lang.ref.WeakReference

/**
 * ゲート画面の起動と、防御策3(自動終了)を担う。
 *
 * 自動終了のタイマーはゲート画面(Activity)の外側、専用のバックグラウンドスレッドで動かす。
 * 画面のUIスレッドが固まっていても、タイマー自体は止まらない。
 *  1. [AUTO_CLOSE_MILLIS] 経過したら、メインスレッドでfinishを試みる
 *  2. それでも [FORCE_CLOSE_GRACE_MILLIS] 後に画面が残っていれば、アクセシビリティの
 *     グローバル操作(ホームに戻る)で強制的に画面から外す(UIスレッドに依存しない)
 * 脱出口の操作の成否とは無関係に働く。
 */
object GateLauncher {
    const val AUTO_CLOSE_MILLIS = 2 * 60_000L
    private const val FORCE_CLOSE_GRACE_MILLIS = 5_000L

    private val timerThread = HandlerThread("gate-auto-close").apply { start() }
    private val timer = Handler(timerThread.looper)
    private val main = Handler(Looper.getMainLooper())

    @Volatile
    private var current: WeakReference<GateActivity>? = null

    @Volatile
    private var service: WeakReference<AccessibilityService>? = null

    val isShowing: Boolean get() = current?.get() != null

    fun bindService(s: AccessibilityService?) {
        service = s?.let { WeakReference(it) }
    }

    fun onGateCreated(activity: GateActivity) {
        current = WeakReference(activity)
    }

    fun onGateDestroyed(activity: GateActivity) {
        if (current?.get() === activity) current = null
    }

    fun launch(context: Context) {
        val intent = Intent(context, GateActivity::class.java).addFlags(
            Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK or Intent.FLAG_ACTIVITY_NO_ANIMATION,
        )
        context.startActivity(intent)
        GuardStore(context).recordGateShown()
        scheduleAutoClose()
    }

    private fun scheduleAutoClose() {
        timer.removeCallbacksAndMessages(null)
        timer.postDelayed({
            val activity = current?.get() ?: return@postDelayed
            main.post { activity.finishAndRemoveTask() }
            timer.postDelayed({
                if (current?.get() != null) {
                    service?.get()?.performGlobalAction(AccessibilityService.GLOBAL_ACTION_HOME)
                }
            }, FORCE_CLOSE_GRACE_MILLIS)
        }, AUTO_CLOSE_MILLIS)
    }
}
