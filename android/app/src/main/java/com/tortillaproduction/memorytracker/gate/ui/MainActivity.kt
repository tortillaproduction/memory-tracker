package com.tortillaproduction.memorytracker.gate.ui

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.lifecycle.lifecycleScope
import com.google.mlkit.vision.barcode.common.Barcode
import com.google.mlkit.vision.codescanner.GmsBarcodeScannerOptions
import com.google.mlkit.vision.codescanner.GmsBarcodeScanning
import com.tortillaproduction.memorytracker.gate.CredentialsStore
import com.tortillaproduction.memorytracker.gate.GateBackend
import com.tortillaproduction.memorytracker.gate.GateLauncher
import com.tortillaproduction.memorytracker.gate.GuardStore
import com.tortillaproduction.memorytracker.gate.SystemSettings
import com.tortillaproduction.memorytracker.gate.TargetStore
import com.tortillaproduction.memorytracker.gate.api.FetchResult
import com.tortillaproduction.memorytracker.gate.api.GateClient
import com.tortillaproduction.memorytracker.gate.api.SetupCode
import com.tortillaproduction.memorytracker.gate.policy.TargetCandidates
import kotlinx.coroutines.launch
import java.time.Instant
import java.time.format.DateTimeFormatter

/**
 * アプリの入口。セットアップが終わっていなければチェックリスト式のセットアップを、
 * 終わっていれば設定画面を出す。設定画面からセットアップをやり直すこともできる。
 *
 * OSの設定画面から戻ったとき(onResume)に、各ステップの完了を自動で判定し直す。
 */
class MainActivity : ComponentActivity() {

    private lateinit var targetStore: TargetStore
    private lateinit var guardStore: GuardStore

    private var inSetup by mutableStateOf(false)
    private var step by mutableStateOf(SetupStep.Connect)
    private var setup by mutableStateOf(SetupState())
    private var settings by mutableStateOf(SettingsState())
    private var message by mutableStateOf<String?>(null)

    /** ④でユーザーがチェックを変えたか。変えていれば、画面を再描画してもその選択を保つ。 */
    private var selectionTouched = false

    /** ⑤を開いた時刻。これ以降にゲートが出たら「出た」と判定する。 */
    private var testStartedAt = 0L

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        targetStore = TargetStore(this)
        guardStore = GuardStore(this)
        inSetup = !targetStore.isSetupCompleted()

        setContent {
            GateTheme {
                if (inSetup) {
                    SetupScreen(
                        state = setup,
                        step = step,
                        message = message,
                        onSelectStep = ::goTo,
                        actions = setupActions,
                    )
                } else {
                    SettingsScreen(state = settings, message = message, actions = settingsActions)
                }
            }
        }

        if (inSetup) {
            lifecycleScope.launch {
                refresh()
                step = SetupStep.entries.firstOrNull { !setup.isDone(it) } ?: SetupStep.Test
            }
        }
    }

    override fun onResume() {
        super.onResume()
        lifecycleScope.launch { refresh() }
    }

    private suspend fun refresh() {
        val connected = CredentialsStore(this).current()
        val candidates = TargetCandidates.select(SystemSettings.launchableApps(this))
        val selected = when {
            selectionTouched -> setup.selected
            targetStore.isConfigured() -> targetStore.targets()
            else -> candidates.map { it.packageName }.toSet() // 初期は候補をすべてオン
        }
        setup = setup.copy(
            connected = connected != null,
            accessibilityOn = SystemSettings.isAccessibilityEnabled(this),
            batteryExempt = SystemSettings.isIgnoringBatteryOptimizations(this),
            appsSaved = targetStore.isConfigured(),
            gateSeen = testStartedAt > 0 && (guardStore.lastGateShownAt() ?: 0) >= testStartedAt,
            candidates = candidates,
            selected = selected,
            maker = SystemSettings.maker,
        )
        val now = System.currentTimeMillis()
        settings = SettingsState(
            serviceEnabled = setup.accessibilityOn,
            server = connected?.baseUrl,
            lastGateShown = guardStore.lastGateShownAt()?.let { format(it) },
            suspendedUntil = guardStore.suspendedUntil().takeIf { it > now }?.let { format(it) },
            targetCount = targetStore.targets().size,
            maker = SystemSettings.maker,
        )
    }

    private fun goTo(next: SetupStep) {
        message = null
        step = next
        if (next == SetupStep.Test) enterTest()
    }

    // --- セットアップ ---

    private val setupActions = SetupActions(
        scanQr = ::scanQr,
        saveSetupCode = ::saveSetupCode,
        openAccessibility = { SystemSettings.openAccessibilitySettings(this) },
        openAppInfo = { SystemSettings.openAppDetails(this) },
        requestBattery = { SystemSettings.requestIgnoreBatteryOptimizations(this) },
        toggleApp = { pkg, on ->
            selectionTouched = true
            setup = setup.copy(selected = if (on) setup.selected + pkg else setup.selected - pkg)
        },
        saveApps = {
            targetStore.save(setup.selected)
            selectionTouched = false
            setup = setup.copy(appsSaved = true)
            goTo(SetupStep.Test)
        },
        launchApp = { pkg ->
            if (!SystemSettings.launchApp(this, pkg)) message = "Could not open the app."
        },
        finish = {
            targetStore.setSetupCompleted(true)
            if (!targetStore.isConfigured()) targetStore.save(setup.selected)
            inSetup = false
            message = null
            lifecycleScope.launch { refresh() }
        },
    )

    /** ①: WebのQRコードを読み取る。カメラ権限は不要(スキャン画面はGoogle Play開発者サービスが出す)。 */
    private fun scanQr() {
        val options = GmsBarcodeScannerOptions.Builder().setBarcodeFormats(Barcode.FORMAT_QR_CODE).build()
        GmsBarcodeScanning.getClient(this, options).startScan()
            .addOnSuccessListener { saveSetupCode(it.rawValue.orEmpty()) }
            .addOnFailureListener { message = "Could not scan. Paste the setup code instead." }
    }

    /** 読み取ったコードを検証してから保存する。無効なトークン(再発行済みなど)は保存しない。 */
    private fun saveSetupCode(raw: String) {
        val credentials = SetupCode.parse(raw)
        if (credentials == null) {
            message = "This is not a Memory Tracker setup code."
            return
        }
        lifecycleScope.launch {
            val check = GateClient(credentials).fetchCandidates()
            if (check == FetchResult.Failed(FetchResult.Reason.Unauthorized)) {
                message = "This code is no longer valid. Create a new QR code on the web."
                return@launch
            }
            CredentialsStore(this@MainActivity).save(credentials)
            message = if (check is FetchResult.Failed) "Saved. Could not reach the server right now." else "Connected."
            refresh()
        }
    }

    /** ⑤: いまゲートが出る状況か(期限切れのサイトがあるか、今日済みでないか)をサーバーに確認する。 */
    private fun enterTest() {
        testStartedAt = System.currentTimeMillis()
        setup = setup.copy(gateSeen = false, serverCheck = "Checking…")
        lifecycleScope.launch {
            val local = guardStore.isFinishedToday()
            val text = when (val r = GateBackend.client(this@MainActivity)?.fetchCandidates()) {
                null -> "Not connected. Go back to step 1."
                is FetchResult.Show ->
                    if (local) {
                        "You opened a site or skipped today, so the gate stays off until tomorrow."
                    } else {
                        "${r.candidates.size} overdue site(s). The gate will show."
                    }
                FetchResult.AlreadyDoneToday -> "You already checked in today, so the gate stays off until tomorrow."
                FetchResult.NoCandidates -> "No overdue sites right now, so the gate won't show."
                is FetchResult.Failed -> "Could not reach the server (${r.reason}), so the gate won't show."
            }
            setup = setup.copy(serverCheck = text)
        }
    }

    // --- 設定画面 ---

    private val settingsActions = SettingsActions(
        tryGate = ::tryGate,
        runSetup = {
            inSetup = true
            message = null
            goTo(SetupStep.Connect)
        },
        openAccessibility = { SystemSettings.openAccessibilitySettings(this) },
        openAppInfo = { SystemSettings.openAppDetails(this) },
        removeSetup = {
            lifecycleScope.launch {
                CredentialsStore(this@MainActivity).clear()
                message = "Disconnected."
                refresh()
            }
        },
    )

    /** 動作確認用: 端末側の制限(今日済み・発動制限)を無視して、サーバーの候補でゲートを出す。 */
    private fun tryGate() {
        lifecycleScope.launch {
            val client = GateBackend.client(this@MainActivity)
            if (client == null) {
                message = "Not connected. Run setup first."
                return@launch
            }
            message = when (val r = client.fetchCandidates()) {
                is FetchResult.Show -> {
                    GateLauncher.launch(this@MainActivity, r.candidates)
                    null
                }
                FetchResult.AlreadyDoneToday -> "Already done today. The gate won't show."
                FetchResult.NoCandidates -> "No overdue sites. The gate won't show."
                is FetchResult.Failed -> "Could not reach the server (${r.reason}). The gate won't show."
            }
        }
    }

    private fun format(millis: Long): String = timeFormat.format(Instant.ofEpochMilli(millis))

    private companion object {
        val timeFormat: DateTimeFormatter =
            DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm").withZone(GuardStore.ZONE)
    }
}
