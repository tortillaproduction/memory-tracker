package com.tortillaproduction.memorytracker.gate.policy

/**
 * 防御策5: ゲートを絶対に出さないパッケージ。ユーザー設定では外せないようコードに固定する。
 *
 * 端末ごとに異なるホーム画面・キーボード・既定の電話アプリは、実行時にも取得して
 * [TriggerPolicy] に渡す(このリストはその取得に失敗しても効く下限)。
 */
object SystemExclusions {

    /**
     * 前面に出ても「アプリの切り替わり」とみなさないパッケージ。
     * 通知シェードやキーボードが一瞬前面になっても、元のアプリに戻ったときに再発動させないため。
     */
    val transparent: Set<String> = setOf(
        "com.android.systemui",
        // キーボード
        "com.google.android.inputmethod.latin",
        "com.android.inputmethod.latin",
        "com.google.android.inputmethod.japanese",
        "com.oplus.inputmethod",
        "com.baidu.input_oppo",
        "com.touchtype.swiftkey",
        "com.motorola.inputmethod",
        "jp.co.omronsoft.openwnn",
        "com.justsystems.atokmobile.service",
        "com.simeji.android.japanese",
    )

    /** 前面になってもゲートを出さないパッケージ([transparent]を含む)。 */
    val excluded: Set<String> = transparent + setOf(
        // OS・システムダイアログ
        "android",
        "com.android.settings",
        "com.android.permissioncontroller",
        "com.google.android.permissioncontroller",
        "com.android.packageinstaller",
        "com.google.android.packageinstaller",
        // 電話・連絡先
        "com.android.phone",
        "com.android.server.telecom",
        "com.android.dialer",
        "com.google.android.dialer",
        "com.android.incallui",
        "com.android.contacts",
        "com.google.android.contacts",
        "com.motorola.dialer",
        "com.oplus.dialer",
        "com.coloros.dialer",
        // 緊急通報・緊急速報
        "com.android.emergency",
        "com.google.android.apps.safetyhub",
        "com.android.cellbroadcastreceiver",
        "com.google.android.cellbroadcastreceiver",
        "com.android.cellbroadcastreceiver.module",
        "com.oplus.sos",
        // ホーム画面
        "com.android.launcher",
        "com.android.launcher3",
        "com.google.android.apps.nexuslauncher",
        "com.motorola.launcher3",
        "com.oppo.launcher",
        "com.coloros.launcher",
        // 省電力・端末管理(OPPO / Motorola)
        "com.coloros.safecenter",
        "com.oplus.battery",
        "com.coloros.phonemanager",
        "com.motorola.ccc.ota",
    )
}
