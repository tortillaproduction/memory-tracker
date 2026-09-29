package com.tortillaproduction.memorytracker.gate.policy

/** 端末にインストールされている、ホーム画面から起動できるアプリ1件。 */
data class AppEntry(val packageName: String, val label: String, val kind: Kind) {
    /** ApplicationInfo.categoryを単純化したもの。 */
    enum class Kind { Social, Video, Game, Other }
}

/**
 * セットアップ④で、ゲートの対象として選べるアプリを決める。
 * SNS・動画・ゲーム系を候補にし、電話・地図・決済・メッセージ系は候補に出さない。
 */
object TargetCandidates {

    /** カテゴリを問わず候補に出さないアプリ(メッセージ・地図・決済・仕事の連絡手段)。 */
    val neverTargets: Set<String> = setOf(
        // メッセージ
        "jp.naver.line.android",
        "com.whatsapp",
        "com.facebook.orca",
        "org.telegram.messenger",
        "com.discord",
        "com.kakao.talk",
        "com.viber.voip",
        "com.skype.raider",
        "com.google.android.apps.messaging",
        "com.android.mms",
        "com.google.android.gm",
        "com.microsoft.teams",
        "us.zoom.videomeetings",
        "com.Slack",
        // 地図・移動
        "com.google.android.apps.maps",
        "com.waze",
        "jp.co.yahoo.android.apps.map",
        "jp.co.yahoo.android.apps.transit",
        "jp.co.jorudan.nrkj",
        // 決済・銀行
        "jp.ne.paypay.android.app",
        "com.google.android.apps.walletnfcrel",
        "jp.co.rakuten.pay",
        "com.paypal.android.p2pmobile",
        "jp.co.ssi.suica",
    )

    fun select(installed: List<AppEntry>): List<AppEntry> = installed
        .asSequence()
        .filter { it.packageName !in SystemExclusions.excluded && it.packageName !in neverTargets }
        .filter {
            it.packageName in TargetApps.defaults ||
                it.kind == AppEntry.Kind.Social ||
                it.kind == AppEntry.Kind.Video ||
                it.kind == AppEntry.Kind.Game
        }
        .distinctBy { it.packageName }
        .sortedBy { it.label.lowercase() }
        .toList()
}
