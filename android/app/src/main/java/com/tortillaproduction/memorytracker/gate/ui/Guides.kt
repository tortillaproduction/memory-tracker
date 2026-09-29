package com.tortillaproduction.memorytracker.gate.ui

import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.tortillaproduction.memorytracker.gate.SystemSettings

@Composable
fun SectionTitle(text: String) {
    Text(text, style = MaterialTheme.typography.titleMedium, modifier = Modifier.padding(top = 12.dp))
}

@Composable
fun Note(text: String) {
    Text(text, style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
}

/**
 * 機種独自の省電力・自動起動の案内(OPPO / Motorola)。メニューの名前はOSのバージョンで変わるため、
 * 「アプリ情報」を開くボタンを添えて、近い項目を探してもらう。
 */
@Composable
fun MakerPowerGuide(maker: SystemSettings.Maker, onOpenAppInfo: () -> Unit) {
    val steps = when (maker) {
        SystemSettings.Maker.Oppo ->
            "OPPO (ColorOS):\n" +
                "• App info > Battery usage > Allow background activity: On\n" +
                "• App info > Allow auto launch: On\n" +
                "• Recent apps: pull down this app's card (or ⋮ > Lock) to lock it"
        SystemSettings.Maker.Motorola ->
            "Motorola:\n" +
                "• App info > App battery usage > Unrestricted\n" +
                "• Settings > Battery > Adaptive battery: keep this app unrestricted"
        SystemSettings.Maker.Other ->
            "• App info > Battery > Unrestricted (or Don't optimize)"
    }
    Note("$steps\n\nMenu names may differ by OS version.")
    OutlinedButton(onClick = onOpenAppInfo, modifier = Modifier.fillMaxWidth()) { Text("Open app info") }
}

/** 防御策6: OSの設定からいつでもゲートを止められることの案内。 */
@Composable
fun StopGuide(onOpenAccessibilitySettings: () -> Unit) {
    SectionTitle("Stop the gate")
    Note(
        "You can turn it off at any time, even if this app stops responding:\n" +
            "Settings > Accessibility > Memory Tracker Gate > Off",
    )
    OutlinedButton(onClick = onOpenAccessibilitySettings, modifier = Modifier.fillMaxWidth()) {
        Text("Open accessibility settings")
    }
}
