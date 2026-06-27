package com.wareg.app

import android.app.Activity
import android.app.TimePickerDialog
import android.os.Bundle
import android.widget.Button
import android.widget.EditText
import android.widget.Switch
import android.widget.Toast
import java.util.Locale

/**
 * Setup / Settings: enter the server URL (saved permanently so you can switch
 * between home IP, Tailscale, etc.), toggle the daily cook reminder, and pick
 * the reminder hour.
 */
class SettingsActivity : Activity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_settings)

        val urlInput = findViewById<EditText>(R.id.url_input)
        val authUserInput = findViewById<EditText>(R.id.auth_user_input)
        val authPassInput = findViewById<EditText>(R.id.auth_pass_input)
        val reminderSwitch = findViewById<Switch>(R.id.reminder_switch)
        val hourButton = findViewById<Button>(R.id.hour_button)
        val save = findViewById<Button>(R.id.save_button)

        urlInput.setText(SettingsStore.getServerUrl(this))
        authUserInput.setText(SettingsStore.getAuthUser(this))
        authPassInput.setText(SettingsStore.getAuthPass(this))
        reminderSwitch.isChecked = SettingsStore.isReminderEnabled(this)
        var hour = SettingsStore.getReminderHour(this)
        hourButton.text = formatHour(hour)

        hourButton.setOnClickListener {
            TimePickerDialog(this, { _, h, _ ->
                hour = h
                hourButton.text = formatHour(hour)
            }, hour, 0, true).show()
        }

        save.setOnClickListener {
            val normalized = SettingsStore.normalize(urlInput.text.toString())
            if (normalized.isEmpty()) {
                Toast.makeText(this, R.string.err_url, Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            SettingsStore.setServerUrl(this, normalized)
            SettingsStore.setAuthUser(this, authUserInput.text.toString().trim())
            SettingsStore.setAuthPass(this, authPassInput.text.toString())
            SettingsStore.setReminderEnabled(this, reminderSwitch.isChecked)
            SettingsStore.setReminderHour(this, hour)
            ReminderScheduler.update(this)
            setResult(Activity.RESULT_OK)
            finish()
        }
    }

    private fun formatHour(hour: Int): String =
        String.format(Locale.US, "Remind at %02d:00", hour)
}
