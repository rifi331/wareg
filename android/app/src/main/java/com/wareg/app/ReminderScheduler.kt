package com.wareg.app

import android.app.AlarmManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.util.Log
import java.util.Calendar

/**
 * Schedules (or cancels) a single daily AlarmManager alarm for the reminder.
 * On fire, [MealPlanReceiver] checks today's meal plan, notifies, and re-arms
 * for the next day.
 */
object ReminderScheduler {
    private const val TAG = "ReminderScheduler"
    const val REQUEST_CODE = 4242

    fun update(ctx: Context) {
        val am = ctx.getSystemService(Context.ALARM_SERVICE) as AlarmManager
        val pi = pendingIntent(ctx)
        am.cancel(pi)

        if (!SettingsStore.isReminderEnabled(ctx)) return
        if (SettingsStore.getServerUrl(ctx).isEmpty()) return

        val hour = SettingsStore.getReminderHour(ctx)
        val cal = Calendar.getInstance().apply {
            set(Calendar.HOUR_OF_DAY, hour)
            set(Calendar.MINUTE, 0)
            set(Calendar.SECOND, 0)
            set(Calendar.MILLISECOND, 0)
            if (timeInMillis <= System.currentTimeMillis()) {
                add(Calendar.DAY_OF_YEAR, 1)
            }
        }

        try {
            am.setExactAndAllowWhileIdle(AlarmManager.RTC_WAKEUP, cal.timeInMillis, pi)
        } catch (e: SecurityException) {
            // Exact alarms may require a user grant on some devices; fall back to inexact.
            Log.w(TAG, "Exact alarm denied, using inexact", e)
            am.setAndAllowWhileIdle(AlarmManager.RTC_WAKEUP, cal.timeInMillis, pi)
        }
    }

    private fun pendingIntent(ctx: Context): PendingIntent {
        val intent = Intent(ctx, MealPlanReceiver::class.java)
        val flags = PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        return PendingIntent.getBroadcast(ctx, REQUEST_CODE, intent, flags)
    }
}
