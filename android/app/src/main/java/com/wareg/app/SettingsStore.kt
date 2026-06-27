package com.wareg.app

import android.content.Context
import android.content.SharedPreferences
import android.util.Base64

/**
 * Persisted app settings (server URL + reminder config + site auth) backed by
 * SharedPreferences. The server URL is stored so you can switch between home
 * IP, a Tailscale URL, etc.
 */
object SettingsStore {
    private const val PREFS = "wareg_prefs"
    private const val KEY_URL = "server_url"
    private const val KEY_REMINDER = "reminder_enabled"
    private const val KEY_HOUR = "reminder_hour"
    private const val KEY_AUTH_USER = "auth_user"
    private const val KEY_AUTH_PASS = "auth_pass"

    private fun prefs(ctx: Context): SharedPreferences =
        ctx.getSharedPreferences(PREFS, Context.MODE_PRIVATE)

    fun getServerUrl(ctx: Context): String =
        prefs(ctx).getString(KEY_URL, "") ?: ""

    fun setServerUrl(ctx: Context, url: String) =
        prefs(ctx).edit().putString(KEY_URL, url).apply()

    fun isReminderEnabled(ctx: Context): Boolean =
        prefs(ctx).getBoolean(KEY_REMINDER, false)

    fun setReminderEnabled(ctx: Context, on: Boolean) =
        prefs(ctx).edit().putBoolean(KEY_REMINDER, on).apply()

    fun getReminderHour(ctx: Context): Int =
        prefs(ctx).getInt(KEY_HOUR, 17)

    fun setReminderHour(ctx: Context, hour: Int) =
        prefs(ctx).edit().putInt(KEY_HOUR, hour).apply()

    fun getAuthUser(ctx: Context): String =
        prefs(ctx).getString(KEY_AUTH_USER, "") ?: ""

    fun setAuthUser(ctx: Context, user: String) =
        prefs(ctx).edit().putString(KEY_AUTH_USER, user).apply()

    fun getAuthPass(ctx: Context): String =
        prefs(ctx).getString(KEY_AUTH_PASS, "") ?: ""

    fun setAuthPass(ctx: Context, pass: String) =
        prefs(ctx).edit().putString(KEY_AUTH_PASS, pass).apply()

    /** Returns "Basic <base64>" if auth is configured, else "". */
    fun getBasicAuthHeader(ctx: Context): String {
        val u = getAuthUser(ctx)
        val p = getAuthPass(ctx)
        if (u.isEmpty() && p.isEmpty()) return ""
        val raw = "$u:$p".toByteArray(Charsets.UTF_8)
        return "Basic " + Base64.encodeToString(raw, Base64.NO_WRAP)
    }

    /** Normalize a user-entered URL: ensure a scheme, strip trailing slashes. */
    fun normalize(input: String): String {
        var s = input.trim()
        if (s.isEmpty()) return ""
        if (!s.contains("://")) s = "http://$s"
        while (s.endsWith("/")) s = s.dropLast(1)
        return s
    }
}
