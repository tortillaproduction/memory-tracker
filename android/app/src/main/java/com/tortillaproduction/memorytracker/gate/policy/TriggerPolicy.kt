package com.tortillaproduction.memorytracker.gate.policy

/**
 * 前面アプリの切り替わりから「ゲートを出す候補か」を判定する。
 * 読むのは前面アプリのpackageNameだけで、画面の内容には触れない。
 *
 * ゲートを出すのは、前面アプリが「前回と違い、かつ対象リストに含まれ、除外リストに含まれない」ときだけ。
 * 同じアプリ内の画面遷移(同じpackageNameのイベントが続く)では出さない。
 * 除外は対象より優先する(防御策5)。
 */
class TriggerPolicy(
    private val ownPackage: String,
    private val targets: Set<String>,
) {
    private var lastForeground: String? = null

    /** 直近の前面アプリ(ゲート画面・通知シェード・キーボードを除く)。 */
    val currentForeground: String? get() = lastForeground

    /**
     * TYPE_WINDOW_STATE_CHANGEDごとに呼ぶ。
     * @param dynamicTransparent 実行時に取得したキーボード(切り替わりとみなさない)
     * @param dynamicExcluded 実行時に取得したホーム画面・既定の電話アプリ
     */
    fun onForegroundChanged(
        packageName: String,
        dynamicTransparent: Set<String> = emptySet(),
        dynamicExcluded: Set<String> = emptySet(),
    ): Boolean {
        // 自分自身(ゲート画面)・通知シェード・キーボードは切り替わりとみなさない。
        // ゲートを閉じて元のアプリに戻ったときに再発動しないようにするため。
        if (packageName == ownPackage ||
            packageName in SystemExclusions.transparent ||
            packageName in dynamicTransparent
        ) {
            return false
        }

        val previous = lastForeground
        lastForeground = packageName

        if (packageName in SystemExclusions.excluded || packageName in dynamicExcluded) {
            return false
        }
        if (packageName == previous) {
            return false
        }
        return packageName in targets
    }
}
