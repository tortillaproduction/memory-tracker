package com.tortillaproduction.memorytracker.gate

/** ゲートに出す期限切れサイト1件(GET /api/gate/candidates の1要素)。 */
data class Candidate(
    val siteId: String,
    val name: String,
    val url: String,
    val overdueHours: Double,
)
