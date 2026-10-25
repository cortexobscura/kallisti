package chat.kallisti.mobile

import android.Manifest
import android.app.Activity
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Build
import app.tauri.annotation.Command
import app.tauri.annotation.InvokeArg
import app.tauri.annotation.TauriPlugin
import app.tauri.plugin.Invoke
import app.tauri.plugin.Plugin

@InvokeArg
class StartArgs {
  var channel: String? = null
}

// Starts and stops VoiceService (see lib.rs: voice_service).
@TauriPlugin
class VoicePlugin(private val activity: Activity) : Plugin(activity) {
  @Command
  fun start(invoke: Invoke) {
    val args = invoke.parseArgs(StartArgs::class.java)
    // Android 13+: without it the service runs, but its notification hides.
    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU &&
      activity.checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS) != PackageManager.PERMISSION_GRANTED
    ) {
      activity.requestPermissions(arrayOf(Manifest.permission.POST_NOTIFICATIONS), 0)
    }
    if (VoiceService.running) {
      VoiceService.update(activity, args.channel ?: "")
      invoke.resolve()
      return
    }
    val intent = Intent(activity, VoiceService::class.java).putExtra(VoiceService.EXTRA_CHANNEL, args.channel ?: "")
    try {
      if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) activity.startForegroundService(intent) else activity.startService(intent)
      invoke.resolve()
    } catch (e: Exception) {
      invoke.reject(e.message ?: "voice service not started")
    }
  }

  @Command
  fun stop(invoke: Invoke) {
    activity.stopService(Intent(activity, VoiceService::class.java))
    invoke.resolve()
  }
}
