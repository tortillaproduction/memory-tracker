package com.tortillaproduction.memorytracker.gate

import android.accessibilityservice.AccessibilityService
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.os.SystemClock
import android.telecom.TelecomManager
import android.view.accessibility.AccessibilityEvent
import android.view.inputmethod.InputMethodManager
import com.tortillaproduction.memorytracker.gate.api.FetchResult
import com.tortillaproduction.memorytracker.gate.policy.RateLimiter
import com.tortillaproduction.memorytracker.gate.policy.StartupGrace
import com.tortillaproduction.memorytracker.gate.policy.TriggerPolicy
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch

/**
 * 前面アプリの切り替わりを検知し、対象アプリならゲート画面を出す。
 *
 * 購読するのはTYPE_WINDOW_STATE_CHANGEDだけで、画面の内容は読まない
 * (canRetrieveWindowContent=false)。使うのはイベントのpackageNameのみ。
 *
 * 防御策6: このサービスはOSの設定(ユーザー補助)からいつでもオフにできる。
 */
class GateAccessibilityService : AccessibilityService() {

    private lateinit var policy: TriggerPolicy
    private lateinit var store: GuardStore
    private lateinit var rateLimiter: RateLimiter
    private var dynamicTransparent: Set<String> = emptySet()
    private var dynamicExcluded: Set<String> = emptySet()
    private var dynamicExclusionsLoadedAt = 0L
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Main.immediate)

    /** 候補の問い合わせ中はtrue。問い合わせを重ねない。 */
    private var fetching = false

    override fun onServiceConnected() {
        super.onServiceConnected()
        val targetStore = TargetStore(this)
        policy = TriggerPolicy(ownPackage = packageName) { targetStore.targets() }
        store = GuardStore(this)
        rateLimiter = store.loadRateLimiter()
        refreshDynamicExclusions(force = true)
        GateLauncher.bindService(this)
    }

    override fun onAccessibilityEvent(event: AccessibilityEvent) {
        if (event.eventType != AccessibilityEvent.TYPE_WINDOW_STATE_CHANGED) return
        val pkg = event.packageName?.toString() ?: return

        refreshDynamicExclusions(force = false)
        if (!policy.onForegroundChanged(pkg, dynamicTransparent, dynamicExcluded)) return

        val now = System.currentTimeMillis()
        if (GateLauncher.isShowing || fetching) return
        if (StartupGrace.isInGrace(SystemClock.elapsedRealtime(), now - lastUpdateTime())) return
        if (store.isFinishedToday(now)) return
        if (!rateLimiter.canAcquire(pkg, now)) return

        fetching = true
        scope.launch {
            try {
                showGateIfNeeded(pkg)
            } finally {
                fetching = false
            }
        }
    }

    /**
     * サーバーに候補を問い合わせ、出すべきときだけゲートを出す。
     * 未設定・通信失敗・タイムアウト(2秒)・候補0件・今日済みのときは何もせず素通しする。
     */
    private suspend fun showGateIfNeeded(pkg: String) {
        val client = GateBackend.client(this) ?: return
        GateBackend.flushPendingDismissal(this)

        when (val result = client.fetchCandidates()) {
            is FetchResult.Show -> {
                // 問い合わせ中にユーザーが別のアプリへ移っていたら出さない
                if (policy.currentForeground != pkg || GateLauncher.isShowing) return
                // 発動制限は実際に出すときに数える(防御策4)
                val allowed = rateLimiter.tryAcquire(pkg, System.currentTimeMillis())
                store.saveRateLimiter(rateLimiter)
                if (allowed) GateLauncher.launch(this, result.candidates)
            }
            FetchResult.AlreadyDoneToday -> store.markDoneToday()
            FetchResult.NoCandidates, is FetchResult.Failed -> Unit
        }
    }

    override fun onInterrupt() = Unit

    override fun onUnbind(intent: Intent?): Boolean {
        GateLauncher.bindService(null)
        return super.onUnbind(intent)
    }

    override fun onDestroy() {
        scope.cancel()
        super.onDestroy()
    }

    private fun lastUpdateTime(): Long = try {
        packageManager.getPackageInfo(packageName, 0).lastUpdateTime
    } catch (e: PackageManager.NameNotFoundException) {
        0L
    }

    /** 端末ごとに異なるホーム画面・キーボード・既定の電話アプリを取得する(防御策5の補完)。 */
    private fun refreshDynamicExclusions(force: Boolean) {
        val now = SystemClock.elapsedRealtime()
        if (!force && now - dynamicExclusionsLoadedAt < 10 * 60_000L) return
        dynamicExclusionsLoadedAt = now

        val keyboards = mutableSetOf<String>()
        val excluded = mutableSetOf<String>()
        runCatching {
            val home = Intent(Intent.ACTION_MAIN).addCategory(Intent.CATEGORY_HOME)
            packageManager.queryIntentActivities(home, 0).forEach { excluded += it.activityInfo.packageName }
        }
        runCatching {
            val imm = getSystemService(Context.INPUT_METHOD_SERVICE) as InputMethodManager
            imm.enabledInputMethodList.forEach { keyboards += it.packageName }
        }
        runCatching {
            val telecom = getSystemService(Context.TELECOM_SERVICE) as TelecomManager
            telecom.defaultDialerPackage?.let { excluded += it }
        }
        dynamicTransparent = keyboards
        dynamicExcluded = excluded
    }
}
