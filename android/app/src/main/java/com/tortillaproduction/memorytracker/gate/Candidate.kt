package com.tortillaproduction.memorytracker.gate

/** ゲートに出す期限切れサイト1件。 */
data class Candidate(
    val siteId: String,
    val name: String,
    val url: String,
    val overdueHours: Double,
)

/** フェーズ3-1用のダミー候補(バックエンドに接続しない動作確認用)。3-2でAPIに置き換える。 */
object DummyCandidates {
    val items = listOf(
        Candidate("dummy-1", "Go by Example", "https://gobyexample.com/", 52.0),
        Candidate("dummy-2", "MDN Web Docs", "https://developer.mozilla.org/", 9.0),
        Candidate("dummy-3", "Kotlin Docs", "https://kotlinlang.org/docs/home.html", 2.0),
    )
}
