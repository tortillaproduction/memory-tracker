package com.tortillaproduction.memorytracker.gate

import android.content.Context
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import com.tortillaproduction.memorytracker.gate.api.Credentials
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map

private val Context.credentialsDataStore by preferencesDataStore(name = "credentials")

/**
 * バックエンドへの接続情報(ベースURLとトークン)をアプリ専用領域に保存する。
 * バックアップは無効(allowBackup=false)にしており、端末外には出ない。
 */
class CredentialsStore(context: Context) {
    private val dataStore = context.applicationContext.credentialsDataStore

    val credentials: Flow<Credentials?> = dataStore.data.map { prefs ->
        val baseUrl = prefs[KEY_BASE_URL]
        val token = prefs[KEY_TOKEN]
        if (baseUrl.isNullOrBlank() || token.isNullOrBlank()) null else Credentials(baseUrl, token)
    }

    suspend fun current(): Credentials? = credentials.first()

    suspend fun save(c: Credentials) {
        dataStore.edit {
            it[KEY_BASE_URL] = c.baseUrl
            it[KEY_TOKEN] = c.token
        }
    }

    suspend fun clear() {
        dataStore.edit { it.clear() }
    }

    private companion object {
        val KEY_BASE_URL = stringPreferencesKey("base_url")
        val KEY_TOKEN = stringPreferencesKey("token")
    }
}
