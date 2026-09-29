package com.tortillaproduction.memorytracker.gate.api

import com.tortillaproduction.memorytracker.gate.Candidate
import kotlinx.coroutines.runBlocking
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import java.util.concurrent.TimeUnit

class GateClientTest {
    private lateinit var server: MockWebServer
    private lateinit var client: GateClient

    @Before
    fun setUp() {
        server = MockWebServer().apply { start() }
        client = GateClient(Credentials(server.url("/").toString(), "test-token"))
    }

    @After
    fun tearDown() {
        server.shutdown()
    }

    private fun json(body: String, code: Int = 200) =
        MockResponse().setResponseCode(code).setHeader("Content-Type", "application/json").setBody(body)

    @Test
    fun `候補があればShowを返し、Bearerトークンを付けて呼ぶ`() = runBlocking {
        server.enqueue(
            json(
                """{"alreadyDoneToday":false,"candidates":[
                {"siteId":"s1","name":"Go","url":"https://go.dev","overdueHours":30.5}]}""",
            ),
        )

        val result = client.fetchCandidates()

        assertEquals(FetchResult.Show(listOf(Candidate("s1", "Go", "https://go.dev", 30.5))), result)
        val req = server.takeRequest()
        assertEquals("/api/gate/candidates", req.path)
        assertEquals("Bearer test-token", req.getHeader("Authorization"))
    }

    @Test
    fun `今日済み・0件なら素通しの結果を返す`() = runBlocking {
        server.enqueue(json("""{"alreadyDoneToday":true,"candidates":[]}"""))
        server.enqueue(json("""{"alreadyDoneToday":false,"candidates":[]}"""))

        assertEquals(FetchResult.AlreadyDoneToday, client.fetchCandidates())
        assertEquals(FetchResult.NoCandidates, client.fetchCandidates())
    }

    @Test
    fun `APIエラー・認証エラー・壊れた応答はFailedになる`() = runBlocking {
        server.enqueue(MockResponse().setResponseCode(500))
        server.enqueue(MockResponse().setResponseCode(401))
        server.enqueue(json("{not json"))

        assertEquals(FetchResult.Failed(FetchResult.Reason.Server), client.fetchCandidates())
        assertEquals(FetchResult.Failed(FetchResult.Reason.Unauthorized), client.fetchCandidates())
        assertEquals(FetchResult.Failed(FetchResult.Reason.InvalidResponse), client.fetchCandidates())
    }

    @Test
    fun `2秒を超える応答はタイムアウトとして打ち切る`() = runBlocking {
        server.enqueue(json("""{"alreadyDoneToday":false,"candidates":[]}""").setHeadersDelay(5, TimeUnit.SECONDS))

        val start = System.currentTimeMillis()
        val result = client.fetchCandidates()
        val elapsed = System.currentTimeMillis() - start

        assertEquals(FetchResult.Failed(FetchResult.Reason.Timeout), result)
        assertTrue("elapsed=$elapsed", elapsed < 3_000)
    }

    @Test
    fun `サーバーに届かなければNetworkとして失敗する`() = runBlocking {
        server.shutdown()
        val result = client.fetchCandidates()
        assertTrue(result is FetchResult.Failed)
    }

    @Test
    fun `チェックインは開くURLを返し、失敗したらnullを返す`() = runBlocking {
        server.enqueue(json("""{"url":"https://go.dev"}"""))
        server.enqueue(MockResponse().setResponseCode(404))

        assertEquals("https://go.dev", client.checkin("s1"))
        val req = server.takeRequest()
        assertEquals("/api/gate/checkin", req.path)
        assertEquals("""{"siteId":"s1"}""", req.body.readUtf8())
        assertNull(client.checkin("other"))
    }

    @Test
    fun `解除の記録は成功ならtrue、失敗ならfalse`() = runBlocking {
        server.enqueue(MockResponse().setResponseCode(204))
        server.enqueue(MockResponse().setResponseCode(500))

        assertTrue(client.dismiss())
        assertEquals("/api/gate/dismiss", server.takeRequest().path)
        assertFalse(client.dismiss())
    }
}
