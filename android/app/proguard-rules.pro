# Keep Kotlin metadata needed by reflection.
-keep class kotlin.Metadata { *; }

# Keep the app's broadcast receivers (instantiated by the system by class name).
-keep class com.wareg.app.MealPlanReceiver { *; }
-keep class com.wareg.app.BootReceiver { *; }
