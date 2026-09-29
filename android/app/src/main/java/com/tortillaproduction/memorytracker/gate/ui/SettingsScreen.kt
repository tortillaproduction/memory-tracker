package com.tortillaproduction.memorytracker.gate.ui

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
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.tortillaproduction.memorytracker.gate.SystemSettings

/** セットアップ後の設定画面に表示する状態。 */
data class SettingsState(
    val serviceEnabled: Boolean = false,
    val server: String? = null,
    val lastGateShown: String? = null,
    val suspendedUntil: String? = null,
    val targetCount: Int = 0,
    val maker: SystemSettings.Maker = SystemSettings.Maker.Other,
)

class SettingsActions(
    val tryGate: () -> Unit,
    val runSetup: () -> Unit,
    val openAccessibility: () -> Unit,
    val openAppInfo: () -> Unit,
    val removeSetup: () -> Unit,
)

/**
 * 設定画面。サービスの稼働状態と最後にゲートが出た日時を表示し、
 * 防御策6(OSの設定から止める手順と設定画面へのボタン)を同じ画面に置く。
 */
@Composable
fun SettingsScreen(state: SettingsState, message: String?, actions: SettingsActions) {
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

            Text("Service: ${if (state.serviceEnabled) "On" else "Off"}")
            Text("Server: ${state.server ?: "Not connected"}")
            Text("Apps: ${state.targetCount}")
            Text("Last gate: ${state.lastGateShown ?: "Never"}")
            state.suspendedUntil?.let { Text("Paused until $it (too many gates)") }

            Button(onClick = actions.tryGate, modifier = Modifier.fillMaxWidth()) { Text("Try the gate now") }
            message?.let { Text(it, color = MaterialTheme.colorScheme.primary) }
            OutlinedButton(onClick = actions.runSetup, modifier = Modifier.fillMaxWidth()) {
                Text("Setup (apps, connection)")
            }

            StopGuide(onOpenAccessibilitySettings = actions.openAccessibility)

            SectionTitle("If the gate stops showing")
            MakerPowerGuide(state.maker, onOpenAppInfo = actions.openAppInfo)

            if (state.server != null) {
                TextButton(onClick = actions.removeSetup) { Text("Disconnect") }
            }
        }
    }
}
