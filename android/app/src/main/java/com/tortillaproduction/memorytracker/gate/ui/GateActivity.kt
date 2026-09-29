package com.tortillaproduction.memorytracker.gate.ui

import android.content.ActivityNotFoundException
import android.content.Intent
import android.net.Uri
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.OnBackPressedCallback
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import com.tortillaproduction.memorytracker.gate.Candidate
import com.tortillaproduction.memorytracker.gate.DummyCandidates
import com.tortillaproduction.memorytracker.gate.GateLauncher
import com.tortillaproduction.memorytracker.gate.GuardStore

/**
 * 対象アプリを開いた直後に出す全画面のゲート。
 * 抜ける手段は「サイトを開く」と「脱出口(長押し)」だけ。戻るボタンは無効にする。
 * そのほか、外側のタイマーによる自動終了(防御策3)と、画面から離れたときの終了で必ず閉じる。
 */
class GateActivity : ComponentActivity() {

    private lateinit var store: GuardStore

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        store = GuardStore(this)
        GateLauncher.onGateCreated(this)

        onBackPressedDispatcher.addCallback(this, object : OnBackPressedCallback(true) {
            override fun handleOnBackPressed() = Unit
        })

        setContent {
            GateTheme {
                GateScreen(
                    candidates = DummyCandidates.items,
                    onOpen = ::open,
                    onSkip = ::skip,
                )
            }
        }
    }

    private fun open(candidate: Candidate) {
        store.markDoneToday()
        try {
            startActivity(
                Intent(Intent.ACTION_VIEW, Uri.parse(candidate.url)).addFlags(Intent.FLAG_ACTIVITY_NEW_TASK),
            )
        } catch (e: ActivityNotFoundException) {
            // ブラウザが無くても、ゲートに閉じ込めない
        }
        finishAndRemoveTask()
    }

    private fun skip() {
        store.markDismissedToday()
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
