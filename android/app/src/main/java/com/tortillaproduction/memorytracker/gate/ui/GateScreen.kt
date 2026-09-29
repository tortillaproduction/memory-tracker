package com.tortillaproduction.memorytracker.gate.ui

import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawingPadding
import androidx.compose.foundation.layout.width
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.tortillaproduction.memorytracker.gate.Candidate
import com.tortillaproduction.memorytracker.gate.policy.overdueLabel
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

/** 脱出口の長押し時間。 */
private const val HOLD_TO_SKIP_MILLIS = 3_000L

@Composable
fun GateScreen(
    candidates: List<Candidate>,
    openingSiteId: String?,
    onOpen: (Candidate) -> Unit,
    onSkip: () -> Unit,
) {
    Surface(modifier = Modifier.fillMaxSize(), color = MaterialTheme.colorScheme.background) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .safeDrawingPadding()
                .padding(horizontal = 20.dp, vertical = 24.dp),
        ) {
            Text("Before you scroll", style = MaterialTheme.typography.headlineSmall)
            Spacer(Modifier.height(4.dp))
            Text(
                "Open one of these sites.",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            Spacer(Modifier.height(24.dp))

            Column(verticalArrangement = Arrangement.spacedBy(12.dp), modifier = Modifier.weight(1f)) {
                candidates.forEach { c ->
                    CandidateCard(
                        candidate = c,
                        opening = openingSiteId == c.siteId,
                        enabled = openingSiteId == null,
                        onOpen = { onOpen(c) },
                    )
                }
            }

            HoldToSkip(onSkip = onSkip, modifier = Modifier.align(Alignment.CenterHorizontally))
        }
    }
}

@Composable
private fun CandidateCard(candidate: Candidate, opening: Boolean, enabled: Boolean, onOpen: () -> Unit) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    candidate.name,
                    style = MaterialTheme.typography.titleMedium,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    overdueLabel(candidate.overdueHours),
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }
            Spacer(Modifier.width(12.dp))
            Button(onClick = onOpen, enabled = enabled) { Text(if (opening) "Opening…" else "Open") }
        }
    }
}

/**
 * 防御策2: 無条件の脱出口。3秒長押しでその日だけゲートを解除する。
 * 小さくても常に見える位置に置き、隠しコマンドにはしない。
 */
@Composable
private fun HoldToSkip(onSkip: () -> Unit, modifier: Modifier = Modifier) {
    var progress by remember { mutableFloatStateOf(0f) }
    val scope = rememberCoroutineScope()

    Box(
        modifier = modifier
            .padding(top = 16.dp)
            .pointerInput(Unit) {
                detectTapGestures(onPress = {
                    val job = scope.launch {
                        val start = System.currentTimeMillis()
                        while (true) {
                            val elapsed = System.currentTimeMillis() - start
                            progress = (elapsed.toFloat() / HOLD_TO_SKIP_MILLIS).coerceAtMost(1f)
                            if (elapsed >= HOLD_TO_SKIP_MILLIS) {
                                onSkip()
                                break
                            }
                            delay(16)
                        }
                    }
                    tryAwaitRelease()
                    job.cancel()
                    progress = 0f
                })
            }
            .padding(12.dp),
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Text(
                "Hold 3s to skip today",
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            Spacer(Modifier.height(6.dp))
            LinearProgressIndicator(progress = { progress }, modifier = Modifier.width(140.dp))
        }
    }
}
