package com.tortillaproduction.memorytracker.gate.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawingPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.Checkbox
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.tortillaproduction.memorytracker.gate.SystemSettings
import com.tortillaproduction.memorytracker.gate.policy.AppEntry

enum class SetupStep(val title: String) {
    Connect("Connect"),
    Accessibility("Turn on the service"),
    Battery("Allow background"),
    Apps("Choose apps"),
    Test("Try it"),
}

/** セットアップの各ステップの状態(画面から見える値だけ)。 */
data class SetupState(
    val connected: Boolean = false,
    val accessibilityOn: Boolean = false,
    val batteryExempt: Boolean = false,
    val appsSaved: Boolean = false,
    val gateSeen: Boolean = false,
    val candidates: List<AppEntry> = emptyList(),
    val selected: Set<String> = emptySet(),
    val serverCheck: String? = null,
    val maker: SystemSettings.Maker = SystemSettings.Maker.Other,
) {
    fun isDone(step: SetupStep): Boolean = when (step) {
        SetupStep.Connect -> connected
        SetupStep.Accessibility -> accessibilityOn
        SetupStep.Battery -> batteryExempt
        SetupStep.Apps -> appsSaved
        SetupStep.Test -> gateSeen
    }
}

class SetupActions(
    val scanQr: () -> Unit,
    val saveSetupCode: (String) -> Unit,
    val openAccessibility: () -> Unit,
    val openAppInfo: () -> Unit,
    val requestBattery: () -> Unit,
    val toggleApp: (String, Boolean) -> Unit,
    val saveApps: () -> Unit,
    val launchApp: (String) -> Unit,
    val finish: () -> Unit,
)

/** チェックリスト式のセットアップ。上にチェックリスト、下に今のステップの1操作だけを出す。 */
@Composable
fun SetupScreen(
    state: SetupState,
    step: SetupStep,
    message: String?,
    onSelectStep: (SetupStep) -> Unit,
    actions: SetupActions,
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
            Text("Set up the gate", style = MaterialTheme.typography.headlineSmall)
            Checklist(state, current = step, onSelect = onSelectStep)
            HorizontalDivider()

            when (step) {
                SetupStep.Connect -> ConnectStep(state, actions, onNext = { onSelectStep(SetupStep.Accessibility) })
                SetupStep.Accessibility -> AccessibilityStep(state, actions, onNext = { onSelectStep(SetupStep.Battery) })
                SetupStep.Battery -> BatteryStep(state, actions, onNext = { onSelectStep(SetupStep.Apps) })
                SetupStep.Apps -> AppsStep(state, actions)
                SetupStep.Test -> TestStep(state, actions)
            }
            message?.let { Text(it, color = MaterialTheme.colorScheme.primary) }
        }
    }
}

@Composable
private fun Checklist(state: SetupState, current: SetupStep, onSelect: (SetupStep) -> Unit) {
    Column {
        SetupStep.entries.forEachIndexed { i, s ->
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .clickable { onSelect(s) }
                    .padding(vertical = 6.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(if (state.isDone(s)) "✓" else "${i + 1}", modifier = Modifier.width(28.dp))
                Text(
                    s.title,
                    fontWeight = if (s == current) FontWeight.Bold else FontWeight.Normal,
                    color = if (s == current) MaterialTheme.colorScheme.onBackground else MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
        }
    }
}

@Composable
private fun NextButton(enabled: Boolean, onNext: () -> Unit) {
    Button(onClick = onNext, enabled = enabled, modifier = Modifier.fillMaxWidth()) { Text("Next") }
}

@Composable
private fun ConnectStep(state: SetupState, actions: SetupActions, onNext: () -> Unit) {
    var showPaste by remember { mutableStateOf(false) }
    var code by remember { mutableStateOf("") }

    SectionTitle("1. Connect")
    Note("On the web, open the menu > Notifications > Gate, then scan the QR code.")
    if (state.connected) {
        Text("✓ Connected")
        NextButton(enabled = true, onNext = onNext)
    } else {
        Button(onClick = actions.scanQr, modifier = Modifier.fillMaxWidth()) { Text("Scan QR code") }
    }

    TextButton(onClick = { showPaste = !showPaste }) { Text("Can't scan? Paste the setup code") }
    if (showPaste) {
        OutlinedTextField(
            value = code,
            onValueChange = { code = it },
            label = { Text("Setup code") },
            modifier = Modifier.fillMaxWidth(),
            minLines = 2,
        )
        OutlinedButton(
            onClick = {
                actions.saveSetupCode(code)
                code = ""
            },
            enabled = code.isNotBlank(),
            modifier = Modifier.fillMaxWidth(),
        ) { Text("Save") }
    }
}

@Composable
private fun AccessibilityStep(state: SetupState, actions: SetupActions, onNext: () -> Unit) {
    SectionTitle("2. Turn on the service")
    Note(
        "Turn on \"Memory Tracker Gate\" in Accessibility. " +
            "It only sees which app is in front. It never reads what's on your screen.",
    )
    if (state.accessibilityOn) {
        Text("✓ Service is on")
        NextButton(enabled = true, onNext = onNext)
    } else {
        Button(onClick = actions.openAccessibility, modifier = Modifier.fillMaxWidth()) {
            Text("Open accessibility settings")
        }
        // Android 13以降、APKを直接入れたアプリはユーザー補助のスイッチが「制限付き設定」で押せない
        Note("Switch greyed out? Open app info > ⋮ (top right) > Allow restricted settings, then try again.")
        OutlinedButton(onClick = actions.openAppInfo, modifier = Modifier.fillMaxWidth()) { Text("Open app info") }
    }
}

@Composable
private fun BatteryStep(state: SetupState, actions: SetupActions, onNext: () -> Unit) {
    SectionTitle("3. Allow background")
    Note("So your phone doesn't stop the gate to save battery.")
    if (state.batteryExempt) {
        Text("✓ Allowed")
    } else {
        Button(onClick = actions.requestBattery, modifier = Modifier.fillMaxWidth()) { Text("Allow") }
    }
    MakerPowerGuide(state.maker, onOpenAppInfo = actions.openAppInfo)
    NextButton(enabled = true, onNext = onNext)
}

@Composable
private fun AppsStep(state: SetupState, actions: SetupActions) {
    SectionTitle("4. Choose apps")
    Note("The gate shows when you open these apps. Phone, maps, payment and messaging apps are never included.")
    if (state.candidates.isEmpty()) {
        Note("No social, video or game apps found.")
    }
    state.candidates.forEach { app ->
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .clickable { actions.toggleApp(app.packageName, app.packageName !in state.selected) },
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Checkbox(
                checked = app.packageName in state.selected,
                onCheckedChange = { actions.toggleApp(app.packageName, it) },
            )
            Text(app.label)
        }
    }
    Button(onClick = actions.saveApps, modifier = Modifier.fillMaxWidth()) { Text("Save") }
}

@Composable
private fun TestStep(state: SetupState, actions: SetupActions) {
    SectionTitle("5. Try it")
    Note("Open one of your apps now. The gate should appear.")
    state.serverCheck?.let { Text(it) }

    state.candidates.filter { it.packageName in state.selected }.forEach { app ->
        OutlinedButton(onClick = { actions.launchApp(app.packageName) }, modifier = Modifier.fillMaxWidth()) {
            Text("Open ${app.label}")
        }
    }

    if (state.gateSeen) {
        Text("✓ The gate appeared.")
    } else {
        Note(
            "No gate?\n" +
                "• Restart your phone, then try again.\n" +
                "• Wait 5 minutes before trying the same app again.\n" +
                "• After you open a site or skip, the gate stays off until tomorrow.",
        )
    }
    Button(onClick = actions.finish, modifier = Modifier.fillMaxWidth()) { Text("Finish") }
}
