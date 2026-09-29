package com.tortillaproduction.memorytracker.gate.policy

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TriggerPolicyTest {
    private val own = "com.tortillaproduction.memorytracker.gate"
    private val instagram = "com.instagram.android"
    private val youtube = "com.google.android.youtube"
    private val launcher = "com.google.android.apps.nexuslauncher"

    private fun policy() = TriggerPolicy(own) { setOf(instagram, youtube) }

    @Test
    fun `対象アプリに切り替わったら発動する`() {
        val p = policy()
        assertFalse(p.onForegroundChanged(launcher))
        assertTrue(p.onForegroundChanged(instagram))
    }

    @Test
    fun `同じアプリ内の画面遷移では発動しない`() {
        val p = policy()
        assertTrue(p.onForegroundChanged(instagram))
        assertFalse(p.onForegroundChanged(instagram))
        assertFalse(p.onForegroundChanged(instagram))
    }

    @Test
    fun `対象外のアプリでは発動しない`() {
        assertFalse(policy().onForegroundChanged("com.example.notes"))
    }

    @Test
    fun `別のアプリを経由して戻ったら再び候補になる`() {
        val p = policy()
        assertTrue(p.onForegroundChanged(instagram))
        assertFalse(p.onForegroundChanged(launcher))
        assertTrue(p.onForegroundChanged(instagram))
    }

    @Test
    fun `ゲート画面・通知シェード・キーボードは切り替わりとみなさない`() {
        val p = policy()
        assertTrue(p.onForegroundChanged(instagram))
        assertFalse(p.onForegroundChanged(own))
        assertFalse(p.onForegroundChanged(instagram)) // ゲートを閉じて戻っても再発動しない
        assertFalse(p.onForegroundChanged("com.android.systemui"))
        assertFalse(p.onForegroundChanged(instagram))
        assertFalse(p.onForegroundChanged("com.example.keyboard", dynamicTransparent = setOf("com.example.keyboard")))
        assertFalse(p.onForegroundChanged(instagram))
    }

    @Test
    fun `除外リストは対象リストより優先する`() {
        val settings = "com.android.settings"
        val dialer = "com.example.dialer"
        val p = TriggerPolicy(own) { setOf(settings, dialer, instagram) }
        assertFalse(p.onForegroundChanged(settings))
        assertFalse(p.onForegroundChanged(dialer, dynamicExcluded = setOf(dialer)))
    }

    @Test
    fun `電話・連絡先・緊急通報・設定・システムUI・ホーム・キーボードが除外リストに含まれる`() {
        listOf(
            "com.google.android.dialer",
            "com.android.contacts",
            "com.android.emergency",
            "com.android.settings",
            "com.android.systemui",
            "com.android.launcher3",
            "com.google.android.inputmethod.latin",
        ).forEach { assertTrue(it, it in SystemExclusions.excluded) }
    }
}
