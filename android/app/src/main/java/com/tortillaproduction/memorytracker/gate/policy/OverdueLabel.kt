package com.tortillaproduction.memorytracker.gate.policy

/** カードに出す超過時間の表記。24時間未満は時間、それ以上は日数。 */
fun overdueLabel(overdueHours: Double): String {
    val hours = overdueHours.toLong().coerceAtLeast(0)
    return if (hours < 24) "${hours}h overdue" else "${hours / 24}d overdue"
}
