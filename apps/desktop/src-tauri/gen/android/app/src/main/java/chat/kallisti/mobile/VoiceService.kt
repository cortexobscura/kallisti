package chat.kallisti.mobile

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.content.Intent
import android.content.pm.ServiceInfo
import android.os.Build
import android.os.IBinder

// While in voice: a foreground service of type microphone, so Android lets
// the app record and play on in the background and with the screen off. Its
// notification ("In voice: #channel") opens the app.
class VoiceService : Service() {
  override fun onBind(intent: Intent?): IBinder? = null

  override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
    val notification = notification(this, intent?.getStringExtra(EXTRA_CHANNEL) ?: "")
    try {
      if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
        startForeground(NOTIFICATION_ID, notification, ServiceInfo.FOREGROUND_SERVICE_TYPE_MICROPHONE)
      } else {
        startForeground(NOTIFICATION_ID, notification)
      }
      running = true
    } catch (e: Exception) {
      // E.g. the microphone not allowed (14+): no service, no crash.
      android.util.Log.w("Kallisti", "voice service not in the foreground", e)
      stopSelf()
    }
    return START_NOT_STICKY
  }

  override fun onDestroy() {
    running = false
    super.onDestroy()
  }

  companion object {
    const val EXTRA_CHANNEL = "channel"
    private const val CHANNEL_ID = "voice"
    private const val NOTIFICATION_ID = 1

    // Started and not yet stopped: a new channel only updates the
    // notification (no new start, which Android allows only in front).
    @Volatile var running = false

    fun update(context: Context, channel: String) {
      context.getSystemService(NotificationManager::class.java).notify(NOTIFICATION_ID, notification(context, channel))
    }

    private fun notification(context: Context, channel: String): Notification {
      val manager = context.getSystemService(NotificationManager::class.java)
      if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
        manager.createNotificationChannel(NotificationChannel(CHANNEL_ID, "Voice", NotificationManager.IMPORTANCE_LOW))
      }
      val open = PendingIntent.getActivity(
        context, 0,
        Intent(context, MainActivity::class.java).addFlags(Intent.FLAG_ACTIVITY_SINGLE_TOP),
        PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT,
      )
      val builder = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) Notification.Builder(context, CHANNEL_ID) else Notification.Builder(context)
      if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) builder.setForegroundServiceBehavior(Notification.FOREGROUND_SERVICE_IMMEDIATE)
      return builder
        .setSmallIcon(android.R.drawable.ic_btn_speak_now)
        .setContentTitle(if (channel.isEmpty()) "In voice" else "In voice: $channel")
        .setContentIntent(open)
        .setOngoing(true)
        .build()
    }
  }
}
