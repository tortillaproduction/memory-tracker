package com.tortillaproduction.memorytracker.gate.policy

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class StartupGraceTest {
    @Test
    fun `起動直後と更新直後の1分間は出さない`() {
        assertTrue(StartupGrace.isInGrace(sinceBootMillis = 30_000, sinceUpdateMillis = 10 * 60_000))
        assertTrue(StartupGrace.isInGrace(sinceBootMillis = 10 * 60_000, sinceUpdateMillis = 30_000))
        assertFalse(StartupGrace.isInGrace(sinceBootMillis = 61_000, sinceUpdateMillis = 61_000))
    }

    @Test
    fun `超過時間は24時間未満なら時間、それ以上は日数で表す`() {
        assertEquals("2h overdue", overdueLabel(2.7))
        assertEquals("23h overdue", overdueLabel(23.9))
        assertEquals("2d overdue", overdueLabel(52.0))
        assertEquals("0h overdue", overdueLabel(-1.0))
    }
}
