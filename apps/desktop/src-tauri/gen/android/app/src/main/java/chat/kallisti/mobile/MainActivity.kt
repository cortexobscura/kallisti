package chat.kallisti.mobile

import android.os.Bundle
import android.view.View
import androidx.activity.enableEdgeToEdge
import androidx.core.graphics.Insets
import androidx.core.view.ViewCompat
import androidx.core.view.WindowInsetsCompat

class MainActivity : TauriActivity() {
  override fun onCreate(savedInstanceState: Bundle?) {
    enableEdgeToEdge()
    super.onCreate(savedInstanceState)
    // Edge to edge (Android 15+) the keyboard no longer shrinks the window:
    // the web view would push the whole page up, the header out of sight.
    // The keyboard's height as padding below it makes the page shorter
    // instead; while it is open, no gesture bar space below the composer.
    ViewCompat.setOnApplyWindowInsetsListener(findViewById<View>(android.R.id.content)) { view, insets ->
      val keyboard = insets.getInsets(WindowInsetsCompat.Type.ime()).bottom
      view.setPadding(0, 0, 0, keyboard)
      if (keyboard == 0) {
        insets
      } else {
        val bars = insets.getInsets(WindowInsetsCompat.Type.navigationBars())
        WindowInsetsCompat.Builder(insets)
          .setInsets(WindowInsetsCompat.Type.navigationBars(), Insets.of(bars.left, bars.top, bars.right, 0))
          .build()
      }
    }
  }
}
