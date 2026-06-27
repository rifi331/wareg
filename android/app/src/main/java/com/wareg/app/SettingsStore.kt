package com.wareg.app

import android.content.Context
import android.content.SharedPreferences

/**
 * Persisted app settings (server URL + reminder config) backed by SharedPreferences.
 * The server URL is stored so you can switch between home IP, a Tailscale URL, etc.
 */
object SettingsStore {
    private const val PREFS = "wareg_prefs"
    private const val KEY_URL = "server_url"
    private const val KEY_REMINDER = "reminder_enabled"
    private const val KEY_HOUR = "reminder_hour"

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

    /** Normalize a user-entered URL: ensure a scheme, strip trailing slashes. */
    fun normalize(input: String): String {
        var s = input.trim()
        if (s.isEmpty()) return ""
        if (!s.contains("://")) s = "http://$s"
        while (s.endsWith("/")) s = s.dropLast(1)
        return s
    }
}
