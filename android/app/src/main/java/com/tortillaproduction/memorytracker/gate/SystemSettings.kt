package com.tortillaproduction.memorytracker.gate

import android.annotation.SuppressLint
import android.content.ActivityNotFoundException
import android.content.ComponentName
import android.content.Context
import android.content.Intent
import android.content.pm.ApplicationInfo
import android.net.Uri
import android.os.Build
import android.os.PowerManager
import android.provider.Settings
import com.tortillaproduction.memorytracker.gate.policy.AppEntry

/** OSの設定画面を開いたり、設定の状態を調べたりする処理をまとめる。 */
object SystemSettings {

    /** OSの設定でこのアクセシビリティサービスが有効になっているか。 */
    fun isAccessibilityEnabled(context: Context): Boolean {
        val enabled = Settings.Secure.getString(
            context.contentResolver,
            Settings.Secure.ENABLED_ACCESSIBILITY_SERVICES,
        ) ?: return false
        val me = ComponentName(context, GateAccessibilityService::class.java)
        return enabled.split(':').any { ComponentName.unflattenFromString(it) == me }
    }

    /** このサービスの設定画面を直接開く。対応していない端末ではユーザー補助の一覧を開く。 */
    fun openAccessibilitySettings(context: Context) {
        val component = ComponentName(context, GateAccessibilityService::class.java).flattenToString()
        val details = Intent("android.settings.ACCESSIBILITY_DETAILS_SETTINGS")
            .putExtra(Intent.EXTRA_COMPONENT_NAME, component)
        if (!start(context, details)) start(context, Intent(Settings.ACTION_ACCESSIBILITY_SETTINGS))
    }

    fun isIgnoringBatteryOptimizations(context: Context): Boolean {
        val pm = context.getSystemService(Context.POWER_SERVICE) as PowerManager
        return pm.isIgnoringBatteryOptimizations(context.packageName)
    }

    /**
     * バッテリー最適化の除外を求めるダイアログを出す。
     * Playストアでは配布しない(APKを直接渡す)ため、このダイアログを使ってよい。
     */
    @SuppressLint("BatteryLife")
    fun requestIgnoreBatteryOptimizations(context: Context) {
        val intent = Intent(Settings.ACTION_REQUEST_IGNORE_BATTERY_OPTIMIZATIONS, Uri.parse("package:${context.packageName}"))
        if (!start(context, intent)) start(context, Intent(Settings.ACTION_IGNORE_BATTERY_OPTIMIZATION_SETTINGS))
    }

    /** このアプリの「アプリ情報」を開く(制限付き設定の許可や、機種ごとの省電力設定に使う)。 */
    fun openAppDetails(context: Context) {
        start(context, Intent(Settings.ACTION_APPLICATION_DETAILS_SETTINGS, Uri.parse("package:${context.packageName}")))
    }

    fun launchApp(context: Context, packageName: String): Boolean {
        val intent = context.packageManager.getLaunchIntentForPackage(packageName) ?: return false
        return start(context, intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK))
    }

    /** ホーム画面から起動できるアプリの一覧。 */
    fun launchableApps(context: Context): List<AppEntry> {
        val pm = context.packageManager
        val launcher = Intent(Intent.ACTION_MAIN).addCategory(Intent.CATEGORY_LAUNCHER)
        return pm.queryIntentActivities(launcher, 0).map { info ->
            val app = info.activityInfo.applicationInfo
            AppEntry(
                packageName = app.packageName,
                label = info.loadLabel(pm).toString(),
                kind = when (app.category) {
                    ApplicationInfo.CATEGORY_SOCIAL -> AppEntry.Kind.Social
                    ApplicationInfo.CATEGORY_VIDEO -> AppEntry.Kind.Video
                    ApplicationInfo.CATEGORY_GAME -> AppEntry.Kind.Game
                    else -> AppEntry.Kind.Other
                },
            )
        }
    }

    fun appLabel(context: Context, packageName: String): String? = runCatching {
        val pm = context.packageManager
        pm.getApplicationLabel(pm.getApplicationInfo(packageName, 0)).toString()
    }.getOrNull()

    /** 機種ごとの省電力設定の案内に使うメーカー。 */
    enum class Maker { Oppo, Motorola, Other }

    val maker: Maker
        get() = when (Build.MANUFACTURER.lowercase()) {
            "oppo", "realme", "oneplus" -> Maker.Oppo
            "motorola" -> Maker.Motorola
            else -> Maker.Other
        }

    private fun start(context: Context, intent: Intent): Boolean = try {
        if (context !is android.app.Activity) intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
        context.startActivity(intent)
        true
    } catch (e: ActivityNotFoundException) {
        false
    } catch (e: SecurityException) {
        false
    }
}
