package com.tortillaproduction.memorytracker.gate.policy

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class RateLimiterTest {
    private val m = RateLimiter.MINUTE

    @Test
    fun `同じアプリは5分空けないと再発動しない`() {
        val r = RateLimiter()
        assertTrue(r.tryAcquire("a", 0))
        assertFalse(r.tryAcquire("a", 4 * m))
        assertTrue(r.tryAcquire("a", 5 * m))
    }

    @Test
    fun `別のアプリは5分以内でも発動できる`() {
        val r = RateLimiter()
        assertTrue(r.tryAcquire("a", 0))
        assertTrue(r.tryAcquire("b", 1_000))
    }

    @Test
    fun `10分間に5回を超えて発動しようとすると1時間停止する`() {
        val r = RateLimiter()
        (1..5).forEach { assertTrue(r.tryAcquire("app$it", it * 1_000L)) }
        assertFalse(r.tryAcquire("app6", 6_000))
        assertTrue(r.isSuspended(6_000))
        assertFalse(r.tryAcquire("app7", 6_000 + 59 * m))
        assertTrue(r.tryAcquire("app7", 6_000 + 60 * m))
    }

    @Test
    fun `10分を過ぎた発動は回数に数えない`() {
        val r = RateLimiter()
        (1..5).forEach { assertTrue(r.tryAcquire("app$it", it * 1_000L)) }
        assertTrue(r.tryAcquire("app6", 11 * m))
        assertFalse(r.isSuspended(11 * m))
    }

    @Test
    fun `状態を保存・復元しても制限が引き継がれる`() {
        val r = RateLimiter()
        assertTrue(r.tryAcquire("a", 0))
        val restored = RateLimiter().apply { restore(r.snapshot()) }
        assertFalse(restored.tryAcquire("a", 1 * m))
        assertEquals(r.snapshot(), restored.snapshot())
    }

    @Test
    fun `canAcquireは状態を変えずに見込みだけを返す`() {
        val r = RateLimiter()
        assertTrue(r.canAcquire("a", 0))
        assertTrue(r.canAcquire("a", 0))
        assertTrue(r.tryAcquire("a", 0))
        assertFalse(r.canAcquire("a", 4 * m))
        assertTrue(r.canAcquire("b", 4 * m))
    }
}
