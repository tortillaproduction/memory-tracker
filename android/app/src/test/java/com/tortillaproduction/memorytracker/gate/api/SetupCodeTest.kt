package com.tortillaproduction.memorytracker.gate.api

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Test

class SetupCodeTest {
    @Test
    fun `WebのQRコードの中身を解釈できる`() {
        val c = SetupCode.parse("""{"v":1,"baseUrl":"https://mt.example.com/","token":"abc"}""")
        assertEquals(Credentials("https://mt.example.com", "abc"), c)
    }

    @Test
    fun `形式が違うものは受け付けない`() {
        assertNull(SetupCode.parse("not json"))
        assertNull(SetupCode.parse("""{"v":2,"baseUrl":"https://mt.example.com","token":"abc"}"""))
        assertNull(SetupCode.parse("""{"v":1,"baseUrl":"https://mt.example.com","token":" "}"""))
        assertNull(SetupCode.parse("""{"v":1,"baseUrl":"ftp://mt.example.com","token":"abc"}"""))
        assertNull(SetupCode.parse("""{"v":1,"baseUrl":"https://mt.example.com"}"""))
    }

    @Test
    fun `文字列化してもトークンは出ない`() {
        assertFalse(Credentials("https://mt.example.com", "secret-token").toString().contains("secret-token"))
    }
}
