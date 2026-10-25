package chat.kallisti.mobile

import android.Manifest
import android.app.Activity
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Build
import android.os.Handler
import android.os.Looper
import app.tauri.annotation.Command
import app.tauri.annotation.InvokeArg
import app.tauri.annotation.TauriPlugin
import app.tauri.plugin.Invoke
import app.tauri.plugin.Plugin

@InvokeArg
class StartArgs {
  var channel: String? = null
}

// Starts and stops VoiceService (see android.rs: voice_service).
@TauriPlugin
class VoicePlugin(private val activity: Activity) : Plugin(activity) {
  private val main = Handler(Looper.getMainLooper())
  private var wanted: String? = null // the channel while in voice
  private var askedNotifications = false

  @Command
  fun start(invoke: Invoke) {
    val channel = invoke.parseArgs(StartArgs::class.java).channel ?: ""
    wanted = channel
    if (VoiceService.running) {
      VoiceService.update(activity, channel)
    } else {
      startWhenAllowed(0)
    }
    invoke.resolve()
  }

  @Command
  fun stop(invoke: Invoke) {
    wanted = null
    main.removeCallbacksAndMessages(null)
    VoiceService.running = false // a start right after starts it again
    activity.stopService(Intent(activity, VoiceService::class.java))
    invoke.resolve()
  }

  // A microphone service may start only with the microphone allowed: on
  // the first join the page's permission request is still open. Waits for
  // it (up to a minute), unless voice ends meanwhile.
  private fun startWhenAllowed(tries: Int) {
    val channel = wanted ?: return
    if (activity.checkSelfPermission(Manifest.permission.RECORD_AUDIO) != PackageManager.PERMISSION_GRANTED) {
      if (tries < 120) main.postDelayed({ startWhenAllowed(tries + 1) }, 500)
      return
    }
    // After the microphone (one permission request at a time), once:
    // without it the service runs, but its notification hides (13+).
    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU && !askedNotifications &&
      activity.checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS) != PackageManager.PERMISSION_GRANTED
    ) {
      askedNotifications = true
      activity.requestPermissions(arrayOf(Manifest.permission.POST_NOTIFICATIONS), 0)
    }
    val intent = Intent(activity, VoiceService::class.java).putExtra(VoiceService.EXTRA_CHANNEL, channel)
    try {
      if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) activity.startForegroundService(intent) else activity.startService(intent)
    } catch (e: Exception) {
      // Not allowed from the background (12+): voice goes on until then.
      android.util.Log.w("Kallisti", "voice service not started", e)
    }
  }
}
