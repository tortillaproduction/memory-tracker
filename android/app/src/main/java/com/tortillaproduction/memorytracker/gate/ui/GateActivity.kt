package com.tortillaproduction.memorytracker.gate.ui

import android.content.ActivityNotFoundException
import android.content.Intent
import android.net.Uri
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.OnBackPressedCallback
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.lifecycle.lifecycleScope
import com.tortillaproduction.memorytracker.gate.Candidate
import com.tortillaproduction.memorytracker.gate.GateBackend
import com.tortillaproduction.memorytracker.gate.GateLauncher
import com.tortillaproduction.memorytracker.gate.GuardStore
import kotlinx.coroutines.launch

/**
 * 対象アプリを開いた直後に出す全画面のゲート。
 * 抜ける手段は「サイトを開く」と「脱出口(長押し)」だけ。戻るボタンは無効にする。
 * そのほか、外側のタイマーによる自動終了(防御策3)と、画面から離れたときの終了で必ず閉じる。
 */
class GateActivity : ComponentActivity() {

    private lateinit var store: GuardStore
    private var openingSiteId by mutableStateOf<String?>(null)

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val candidates = intent.getCandidates()
        if (candidates.isEmpty()) {
            finishAndRemoveTask()
            return
        }

        enableEdgeToEdge()
        store = GuardStore(this)
        GateLauncher.onGateCreated(this)

        onBackPressedDispatcher.addCallback(this, object : OnBackPressedCallback(true) {
            override fun handleOnBackPressed() = Unit
        })

        setContent {
            GateTheme {
                GateScreen(
                    candidates = candidates,
                    openingSiteId = openingSiteId,
                    onOpen = ::open,
                    onSkip = ::skip,
                )
            }
        }
    }

    /**
     * チェックインを記録してからサイトを開く。記録に失敗(タイムアウト2秒を含む)しても、
     * 候補のURLをそのまま開いてゲートを閉じる(正規の手順の失敗で閉じ込めない)。
     */
    private fun open(candidate: Candidate) {
        if (openingSiteId != null) return
        openingSiteId = candidate.siteId
        store.markDoneToday()

        lifecycleScope.launch {
            val url = GateBackend.client(this@GateActivity)?.checkin(candidate.siteId) ?: candidate.url
            try {
                startActivity(Intent(Intent.ACTION_VIEW, Uri.parse(url)).addFlags(Intent.FLAG_ACTIVITY_NEW_TASK))
            } catch (e: ActivityNotFoundException) {
                // ブラウザが無くても、ゲートに閉じ込めない
            }
            finishAndRemoveTask()
        }
    }

    /** 防御策2: 端末に記録してから閉じ、サーバーへの送信はバックグラウンドで行う。 */
    private fun skip() {
        store.markDismissedToday()
        GateBackend.sendDismissal(this)
        finishAndRemoveTask()
    }

    override fun onStop() {
        super.onStop()
        // ホームボタン等で画面から離れたら、裏に残さず終了する
        if (!isChangingConfigurations) finishAndRemoveTask()
    }

    override fun onDestroy() {
        GateLauncher.onGateDestroyed(this)
        super.onDestroy()
    }
}

private const val EXTRA_IDS = "candidate_ids"
private const val EXTRA_NAMES = "candidate_names"
private const val EXTRA_URLS = "candidate_urls"
private const val EXTRA_HOURS = "candidate_overdue_hours"

fun Intent.putCandidates(candidates: List<Candidate>): Intent = this
    .putExtra(EXTRA_IDS, candidates.map { it.siteId }.toTypedArray())
    .putExtra(EXTRA_NAMES, candidates.map { it.name }.toTypedArray())
    .putExtra(EXTRA_URLS, candidates.map { it.url }.toTypedArray())
    .putExtra(EXTRA_HOURS, candidates.map { it.overdueHours }.toDoubleArray())

private fun Intent.getCandidates(): List<Candidate> {
    val ids = getStringArrayExtra(EXTRA_IDS) ?: return emptyList()
    val names = getStringArrayExtra(EXTRA_NAMES) ?: return emptyList()
    val urls = getStringArrayExtra(EXTRA_URLS) ?: return emptyList()
    val hours = getDoubleArrayExtra(EXTRA_HOURS) ?: return emptyList()
    if (setOf(ids.size, names.size, urls.size, hours.size).size != 1) return emptyList()
    return ids.indices.map { Candidate(ids[it], names[it], urls[it], hours[it]) }
}
