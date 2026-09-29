package com.tortillaproduction.memorytracker.gate.policy

import org.junit.Assert.assertEquals
import org.junit.Test

class TargetCandidatesTest {
    private fun app(pkg: String, label: String, kind: AppEntry.Kind = AppEntry.Kind.Other) = AppEntry(pkg, label, kind)

    @Test
    fun `SNS・動画・ゲーム系と既知の対象アプリを候補にし、名前順に並べる`() {
        val result = TargetCandidates.select(
            listOf(
                app("com.example.game", "zGame", AppEntry.Kind.Game),
                app("com.instagram.android", "Instagram"), // カテゴリ不明でも既知の対象
                app("com.example.video", "Anime Video", AppEntry.Kind.Video),
                app("com.example.notes", "Notes"),
            ),
        )
        assertEquals(listOf("Anime Video", "Instagram", "zGame"), result.map { it.label })
    }

    @Test
    fun `電話・地図・決済・メッセージ系はカテゴリがSNSでも候補に出さない`() {
        val result = TargetCandidates.select(
            listOf(
                app("jp.naver.line.android", "LINE", AppEntry.Kind.Social),
                app("com.google.android.apps.maps", "Maps", AppEntry.Kind.Social),
                app("jp.ne.paypay.android.app", "PayPay", AppEntry.Kind.Social),
                app("com.google.android.dialer", "Phone", AppEntry.Kind.Social),
                app("com.android.settings", "Settings", AppEntry.Kind.Game),
            ),
        )
        assertEquals(emptyList<AppEntry>(), result)
    }
}
