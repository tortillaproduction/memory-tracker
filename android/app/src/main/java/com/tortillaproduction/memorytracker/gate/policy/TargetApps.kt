package com.tortillaproduction.memorytracker.gate.policy

/**
 * ゲートを出す対象アプリ(フェーズ3-1では固定値。3-3でユーザーが選べるようにする)。
 * SNS・動画系のみ。電話・地図・決済・メッセージ系は含めない。
 */
object TargetApps {
    val defaults: Set<String> = setOf(
        "com.instagram.android",
        "com.instagram.barcelona", // Threads
        "com.twitter.android", // X
        "com.zhiliaoapp.musically", // TikTok
        "com.ss.android.ugc.trill", // TikTok(一部地域)
        "com.google.android.youtube",
        "com.facebook.katana",
        "com.reddit.frontpage",
        "com.snapchat.android",
        "tv.twitch.android.app",
    )
}
