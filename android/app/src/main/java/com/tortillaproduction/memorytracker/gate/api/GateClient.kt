package com.tortillaproduction.memorytracker.gate.api

import com.tortillaproduction.memorytracker.gate.Candidate
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import retrofit2.HttpException
import retrofit2.Response
import retrofit2.Retrofit
import retrofit2.converter.kotlinx.serialization.asConverterFactory
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST
import java.io.InterruptedIOException
import java.util.concurrent.TimeUnit

/** 候補取得の結果。ゲートを出すのは [Show] のときだけで、それ以外は素通しする。 */
sealed interface FetchResult {
    data class Show(val candidates: List<Candidate>) : FetchResult
    data object AlreadyDoneToday : FetchResult
    data object NoCandidates : FetchResult
    data class Failed(val reason: Reason) : FetchResult

    enum class Reason { Timeout, Network, Unauthorized, Server, InvalidResponse }
}

/**
 * ゲートAPIのクライアント。どの呼び出しも [timeoutMillis] (既定2秒)で打ち切り、
 * 失敗しても例外を投げずに「素通し」側の結果を返す(ゲートに閉じ込めないため)。
 * Authorizationヘッダーのトークンはログに出さない(HTTPログは有効にしない)。
 */
class GateClient(
    credentials: Credentials,
    timeoutMillis: Long = DEFAULT_TIMEOUT_MILLIS,
) {
    private val api: Api

    init {
        val http = OkHttpClient.Builder()
            .callTimeout(timeoutMillis, TimeUnit.MILLISECONDS)
            .addInterceptor { chain ->
                chain.proceed(
                    chain.request().newBuilder()
                        .header("Authorization", "Bearer ${credentials.token}")
                        .build(),
                )
            }
            .build()
        api = Retrofit.Builder()
            .baseUrl(credentials.baseUrl.trimEnd('/') + "/")
            .client(http)
            .addConverterFactory(json.asConverterFactory("application/json".toMediaType()))
            .build()
            .create(Api::class.java)
    }

    suspend fun fetchCandidates(): FetchResult = try {
        val res = api.candidates()
        when {
            res.alreadyDoneToday -> FetchResult.AlreadyDoneToday
            res.candidates.isEmpty() -> FetchResult.NoCandidates
            else -> FetchResult.Show(
                res.candidates.map { Candidate(it.siteId, it.name, it.url, it.overdueHours) },
            )
        }
    } catch (e: Exception) {
        FetchResult.Failed(classify(e))
    }

    /** チェックインを記録し、開くURLを返す。失敗したらnull(呼び出し側は候補のURLをそのまま開く)。 */
    suspend fun checkin(siteId: String): String? = try {
        api.checkin(CheckinRequest(siteId)).url
    } catch (e: Exception) {
        null
    }

    /** 脱出口で解除したことを記録する。成功したらtrue。 */
    suspend fun dismiss(): Boolean = try {
        api.dismiss().isSuccessful
    } catch (e: Exception) {
        false
    }

    private fun classify(e: Exception): FetchResult.Reason = when (e) {
        is HttpException -> if (e.code() == 401) FetchResult.Reason.Unauthorized else FetchResult.Reason.Server
        is InterruptedIOException -> FetchResult.Reason.Timeout
        is java.io.IOException -> FetchResult.Reason.Network
        else -> FetchResult.Reason.InvalidResponse
    }

    private interface Api {
        @GET("api/gate/candidates")
        suspend fun candidates(): CandidatesResponse

        @POST("api/gate/checkin")
        suspend fun checkin(@Body body: CheckinRequest): CheckinResponse

        @POST("api/gate/dismiss")
        suspend fun dismiss(): Response<Unit>
    }

    @Serializable
    private data class CandidatesResponse(val alreadyDoneToday: Boolean, val candidates: List<CandidateDto>)

    @Serializable
    private data class CandidateDto(val siteId: String, val name: String, val url: String, val overdueHours: Double)

    @Serializable
    private data class CheckinRequest(val siteId: String)

    @Serializable
    private data class CheckinResponse(val url: String)

    companion object {
        const val DEFAULT_TIMEOUT_MILLIS = 2_000L
        private val json = Json { ignoreUnknownKeys = true }
    }
}
