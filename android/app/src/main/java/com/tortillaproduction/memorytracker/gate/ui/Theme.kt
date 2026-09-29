package com.tortillaproduction.memorytracker.gate.ui

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

// Web版と同じ紫系の配色にする。
private val colors = darkColorScheme(
    primary = Color(0xFFA78BFA),
    onPrimary = Color(0xFF1B0F30),
    background = Color(0xFF1B0F30),
    onBackground = Color(0xFFF3EFFF),
    surface = Color(0xFF2A1A47),
    onSurface = Color(0xFFF3EFFF),
    onSurfaceVariant = Color(0xFFC9BDE6),
)

@Composable
fun GateTheme(content: @Composable () -> Unit) {
    MaterialTheme(colorScheme = colors, content = content)
}
