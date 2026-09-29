package com.tortillaproduction.memorytracker.gate

import android.content.Context
import com.tortillaproduction.memorytracker.gate.policy.TargetApps

/**
 * ゲートの対象アプリと、セットアップの進み具合を保存する。
 * アクセシビリティサービスがイベントのたびに同期的に読むため、SharedPreferencesに置く。
 */
class TargetStore(context: Context) {
    private val prefs = context.applicationContext.getSharedPreferences("gate_targets", Context.MODE_PRIVATE)

    /** ユーザーが選んだ対象アプリ。未選択(セットアップ④の前)は既定のSNS・動画アプリ。 */
    fun targets(): Set<String> = prefs.getStringSet(KEY_TARGETS, null) ?: TargetApps.defaults

    fun isConfigured(): Boolean = prefs.contains(KEY_TARGETS)

    fun save(targets: Set<String>) = prefs.edit().putStringSet(KEY_TARGETS, targets.toSet()).apply()

    fun isSetupCompleted(): Boolean = prefs.getBoolean(KEY_SETUP_COMPLETED, false)

    fun setSetupCompleted(v: Boolean) = prefs.edit().putBoolean(KEY_SETUP_COMPLETED, v).apply()

    private companion object {
        const val KEY_TARGETS = "targets"
        const val KEY_SETUP_COMPLETED = "setup_completed"
    }
}
