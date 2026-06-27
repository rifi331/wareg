package com.wareg.app

import android.Manifest
import android.annotation.SuppressLint
import android.app.Activity
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import android.view.Menu
import android.view.MenuItem
import android.webkit.WebResourceRequest
import android.webkit.WebView
import android.webkit.WebViewClient

/**
 * Hosts the Wareg web UI in a WebView. If no server URL is configured yet
 * (first launch), it opens [SettingsActivity] first. The URL is re-read on
 * every resume so changing it (home IP <-> Tailscale) takes effect immediately.
 */
class MainActivity : Activity() {

    private lateinit var web: WebView
    private var loadedUrl: String? = null

    companion object {
        private const val REQ_NOTIF = 11
    }

    @SuppressLint("SetJavaScriptEnabled")
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        if (SettingsStore.getServerUrl(this).isEmpty()) {
            startActivity(Intent(this, SettingsActivity::class.java))
            finish()
            return
        }

        requestNotificationPermissionIfNeeded()

        web = findViewById(R.id.webview)
        with(web.settings) {
            javaScriptEnabled = true          // HTMX + Tailwind CDN need JS
            domStorageEnabled = true
            databaseEnabled = true
            setSupportZoom(true)
            builtInZoomControls = true
            displayZoomControls = false
            loadWithOverviewMode = true
            useWideViewPort = true
        }
        web.webViewClient = object : WebViewClient() {
            override fun shouldOverrideUrlLoading(view: WebView, request: WebResourceRequest): Boolean {
                val u = request.url
                val scheme = u.scheme ?: return false
                // Keep http(s) inside the app; hand other schemes to the system.
                if (scheme == "http" || scheme == "https") return false
                runCatching { startActivity(Intent(Intent.ACTION_VIEW, u)) }
                return true
            }
        }
    }

    private fun requestNotificationPermissionIfNeeded() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU &&
            checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS) != PackageManager.PERMISSION_GRANTED
        ) {
            requestPermissions(arrayOf(Manifest.permission.POST_NOTIFICATIONS), REQ_NOTIF)
        }
    }

    override fun onResume() {
        super.onResume()
        val url = SettingsStore.getServerUrl(this)
        if (url.isNotEmpty() && url != loadedUrl) {
            web.loadUrl("$url/")
            loadedUrl = url
        }
    }

    override fun onCreateOptionsMenu(menu: Menu): Boolean {
        menuInflater.inflate(R.menu.main_menu, menu)
        return true
    }

    override fun onOptionsItemSelected(item: MenuItem): Boolean {
        return when (item.itemId) {
            R.id.action_settings -> {
                startActivity(Intent(this, SettingsActivity::class.java))
                true
            }
            R.id.action_reload -> {
                web.reload()
                true
            }
            else -> super.onOptionsItemSelected(item)
        }
    }

    @Deprecated("Deprecated in Java")
    override fun onBackPressed() {
        @Suppress("DEPRECATION")
        if (this::web.isInitialized && web.canGoBack()) web.goBack() else super.onBackPressed()
    }
}
