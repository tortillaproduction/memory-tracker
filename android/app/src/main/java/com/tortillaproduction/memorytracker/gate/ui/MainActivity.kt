package com.tortillaproduction.memorytracker.gate.ui

import android.content.ComponentName
import android.content.Context
import android.content.Intent
import android.os.Bundle
import android.provider.Settings
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawingPadding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.TextButton
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.lifecycleScope
import com.tortillaproduction.memorytracker.gate.CredentialsStore
import com.tortillaproduction.memorytracker.gate.GateAccessibilityService
import com.tortillaproduction.memorytracker.gate.GateBackend
import com.tortillaproduction.memorytracker.gate.GateLauncher
import com.tortillaproduction.memorytracker.gate.GuardStore
import com.tortillaproduction.memorytracker.gate.api.FetchResult
import com.tortillaproduction.memorytracker.gate.api.SetupCode
import kotlinx.coroutines.launch
import java.time.Instant
import java.time.format.DateTimeFormatter

/**
 * 設定画面(フェーズ3-1の最小構成)。サービスの稼働状態、最後にゲートが出た日時、
 * 防御策6(OSの設定からサービスを止める手順)を表示する。
 * 3-3でチェックイン式のセットアップ画面に置き換える。
 */
class MainActivity : ComponentActivity() {

    private var status by mutableStateOf(Status())
    private var message by mutableStateOf<String?>(null)

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            GateTheme {
                SettingsScreen(
                    status = status,
                    message = message,
                    onSaveSetupCode = ::saveSetupCode,
                    onRemoveSetup = ::removeSetup,
                    onOpenAccessibilitySettings = {
                        startActivity(Intent(Settings.ACTION_ACCESSIBILITY_SETTINGS))
                    },
                    onTryGate = ::tryGate,
                )
            }
        }
    }

    override fun onResume() {
        super.onResume()
        refresh()
    }

    private fun refresh() {
        lifecycleScope.launch { status = loadStatus(this@MainActivity) }
    }

    private fun saveSetupCode(raw: String) {
        val credentials = SetupCode.parse(raw)
        if (credentials == null) {
            message = "Invalid setup code."
            return
        }
        lifecycleScope.launch {
            CredentialsStore(this@MainActivity).save(credentials)
            message = "Connected."
            refresh()
        }
    }

    private fun removeSetup() {
        lifecycleScope.launch {
            CredentialsStore(this@MainActivity).clear()
            message = "Setup removed."
            refresh()
        }
    }

    /** 動作確認用: 端末側の制限(今日済み・発動制限)を無視して、サーバーの候補でゲートを出す。 */
    private fun tryGate() {
        lifecycleScope.launch {
            val client = GateBackend.client(this@MainActivity)
            if (client == null) {
                message = "Not connected. Save a setup code first."
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
}

data class Status(
    val serviceEnabled: Boolean = false,
    val server: String? = null,
    val lastGateShown: String? = null,
    val suspendedUntil: String? = null,
)

private val timeFormat: DateTimeFormatter =
    DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm").withZone(GuardStore.ZONE)

private suspend fun loadStatus(context: Context): Status {
    val store = GuardStore(context)
    val now = System.currentTimeMillis()
    return Status(
        serviceEnabled = isServiceEnabled(context),
        server = CredentialsStore(context).current()?.baseUrl,
        lastGateShown = store.lastGateShownAt()?.let { timeFormat.format(Instant.ofEpochMilli(it)) },
        suspendedUntil = store.suspendedUntil().takeIf { it > now }?.let { timeFormat.format(Instant.ofEpochMilli(it)) },
    )
}

/** OSの設定でこのアクセシビリティサービスが有効になっているか。 */
fun isServiceEnabled(context: Context): Boolean {
    val enabled = Settings.Secure.getString(
        context.contentResolver,
        Settings.Secure.ENABLED_ACCESSIBILITY_SERVICES,
    ) ?: return false
    val me = ComponentName(context, GateAccessibilityService::class.java)
    return enabled.split(':').any { ComponentName.unflattenFromString(it) == me }
}

@Composable
private fun SettingsScreen(
    status: Status,
    message: String?,
    onSaveSetupCode: (String) -> Unit,
    onRemoveSetup: () -> Unit,
    onOpenAccessibilitySettings: () -> Unit,
    onTryGate: () -> Unit,
) {
    Surface(modifier = Modifier.fillMaxSize(), color = MaterialTheme.colorScheme.background) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .safeDrawingPadding()
                .verticalScroll(rememberScrollState())
                .padding(20.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Text("Memory Tracker Gate", style = MaterialTheme.typography.headlineSmall)

            Text("Service: ${if (status.serviceEnabled) "On" else "Off"}")
            Text("Server: ${status.server ?: "Not connected"}")
            Text("Last gate: ${status.lastGateShown ?: "Never"}")
            status.suspendedUntil?.let { Text("Paused until $it (too many gates)") }

            Button(onClick = onTryGate, modifier = Modifier.fillMaxWidth()) {
                Text("Try the gate now")
            }
            message?.let { Text(it, color = MaterialTheme.colorScheme.primary) }

            SetupCodeSection(connected = status.server != null, onSave = onSaveSetupCode, onRemove = onRemoveSetup)

            Text("Stop the gate", style = MaterialTheme.typography.titleMedium, modifier = Modifier.padding(top = 12.dp))
            Text(
                "You can turn it off at any time:\n" +
                    "Settings > Accessibility > Memory Tracker Gate > Off",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            OutlinedButton(onClick = onOpenAccessibilitySettings, modifier = Modifier.fillMaxWidth()) {
                Text("Open accessibility settings")
            }
        }
    }
}

/** QRコードの代わりにセットアップコード(QRの中身)を貼り付けて接続する。 */
@Composable
private fun SetupCodeSection(connected: Boolean, onSave: (String) -> Unit, onRemove: () -> Unit) {
    var code by remember { mutableStateOf("") }
    Text("Setup code", style = MaterialTheme.typography.titleMedium, modifier = Modifier.padding(top = 12.dp))
    OutlinedTextField(
        value = code,
        onValueChange = { code = it },
        label = { Text("Paste the setup code") },
        modifier = Modifier.fillMaxWidth(),
        minLines = 2,
    )
    Button(
        onClick = {
            onSave(code)
            code = ""
        },
        enabled = code.isNotBlank(),
        modifier = Modifier.fillMaxWidth(),
    ) { Text("Save") }
    if (connected) {
        TextButton(onClick = onRemove) { Text("Remove setup") }
    }
}
