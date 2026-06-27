package com.wareg.app

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.os.Build
import org.json.JSONArray
import java.net.HttpURLConnection
import java.net.URL
import kotlin.concurrent.thread

/**
 * Fired by [ReminderScheduler] at the configured hour. Fetches today's meal
 * plan from the configured server and posts a notification listing the meals.
 * If nothing is planned today, no notification is shown. Always re-arms itself.
 */
class MealPlanReceiver : BroadcastReceiver() {

    companion object {
        const val CHANNEL_ID = "wareg_meal_reminder"
        const val NOTIF_ID = 1001
    }

    override fun onReceive(context: Context, intent: Intent) {
        val pending = goAsync()
        thread {
            try {
                handle(context)
            } finally {
                pending.finish()
            }
        }
    }

    private fun handle(ctx: Context) {
        if (!SettingsStore.isReminderEnabled(ctx)) return
        val base = SettingsStore.getServerUrl(ctx)
        if (base.isEmpty()) return

        val titles = fetchTodayMeals(ctx, base)
        if (titles.isNotEmpty()) {
            ensureChannel(ctx)
            notify(ctx, titles)
        }
        // Re-arm for tomorrow regardless.
        ReminderScheduler.update(ctx)
    }

    private fun fetchTodayMeals(ctx: Context, base: String): List<String> {
        val out = ArrayList<String>()
        var conn: HttpURLConnection? = null
        try {
            conn = (URL("$base/api/meal-plan/today").openConnection() as HttpURLConnection).apply {
                connectTimeout = 10_000
                readTimeout = 10_000
                useCaches = false
                setRequestProperty("Accept", "application/json")
                val auth = SettingsStore.getBasicAuthHeader(ctx)
                if (auth.isNotEmpty()) setRequestProperty("Authorization", auth)
            }
            if (conn.responseCode in 200..299) {
                val body = conn.inputStream.bufferedReader().use { it.readText() }
                val arr = JSONArray(body)
                for (i in 0 until arr.length()) {
                    val title = arr.optJSONObject(i)?.optString("recipe_title")
                    if (!title.isNullOrEmpty()) out.add(title)
                }
            }
        } catch (_: Exception) {
            // Network error / server down: stay silent.
        } finally {
            conn?.disconnect()
        }
        return out
    }

    private fun ensureChannel(ctx: Context) {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val mgr = ctx.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
            if (mgr.getNotificationChannel(CHANNEL_ID) == null) {
                val ch = NotificationChannel(
                    CHANNEL_ID,
                    "Cook reminders",
                    NotificationManager.IMPORTANCE_DEFAULT
                ).apply { description = "Daily reminders to cook your planned meals" }
                mgr.createNotificationChannel(ch)
            }
        }
    }

    private fun notify(ctx: Context, titles: List<String>) {
        val mgr = ctx.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
        val text = titles.joinToString(", ")
        val openIntent = Intent(ctx, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        val pi = PendingIntent.getActivity(
            ctx, 0, openIntent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
        val n = Notification.Builder(ctx, CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_notification)
            .setContentTitle("Time to cook!")
            .setContentText(text)
            .setStyle(Notification.BigTextStyle().bigText(text))
            .setContentIntent(pi)
            .setAutoCancel(true)
            .build()
        mgr.notify(NOTIF_ID, n)
    }
}
