# Wareg Android app

A thin Android client that loads the Wareg web UI from a server URL you choose
(home IP, a Tailscale URL, etc.), and sends a **daily cook reminder** listing
that day's planned meals.

- **First launch** opens a setup screen to enter the server URL (saved
  permanently — change it later any time from Settings).
- **Main screen** is the full Wareg UI in a WebView (identical to the web app).
- **Settings** (overflow menu → *Settings*): server URL, reminder on/off,
  reminder hour.
- At the chosen hour the app fetches `GET <serverURL>/api/meal-plan/today`; if
  meals are planned today it posts a notification ("Time to cook: …"). If the
  reminder is off, nothing happens. The alarm re-arms itself daily and after
  reboot.

## Build the APK (Android Studio)

1. Install **Android Studio** (Hedgehog/2023.1.1 or newer, which bundles
   JDK 17 + Android SDK).
2. **File → Open** and select this `android/` folder.
3. When prompted, let Android Studio generate the **Gradle wrapper**
   (`gradlew` / `gradle-wrapper.jar`) and download the SDK/Gradle it asks for.
   - If it does not auto-generate the wrapper, run once in the terminal:
     `gradle wrapper` (needs a local Gradle) — or just use the IDE's bundled
     Gradle.
4. Wait for Gradle sync to finish.
5. **Build → Build Bundle(s) / APK(s) → Build APK(s)**.
6. The APK appears at `app/build/outputs/apk/debug/app-debug.apk`. Use
   `Run ▶` to install on a connected device/emulator, or copy the APK to your
   phone and install it (enable "Install unknown apps").

> This build is a **debug-signed** APK (fine for personal/Tailscale use). For a
> release-signed build use **Build → Generate Signed Bundle / APK**.

## Configuring

- **Server URL**: e.g. `192.168.100.176:7001` or your Tailscale URL. `http://`
  is added automatically if you omit the scheme; cleartext (http) is allowed so
  LAN/Tailscale works.
- **Reminder**: turn the switch on and pick the hour (local device time). The
  Wareg server must be reachable from the phone at that time for the
  notification to list meals.

## Files

```
android/
├── settings.gradle, build.gradle, gradle.properties      # Gradle config
├── gradle/wrapper/gradle-wrapper.properties              # Gradle 8.4
└── app/
    ├── build.gradle, proguard-rules.pro
    └── src/main/
        ├── AndroidManifest.xml
        ├── java/com/wareg/app/
        │   ├── MainActivity.kt          # WebView host
        │   ├── SettingsActivity.kt      # URL + reminder setup
        │   ├── SettingsStore.kt         # persisted settings
        │   ├── ReminderScheduler.kt     # daily AlarmManager
        │   ├── MealPlanReceiver.kt      # fetch today's plan + notify
        │   └── BootReceiver.kt          # re-arm after reboot
        └── res/                         # layouts, strings, themes, icons
```
