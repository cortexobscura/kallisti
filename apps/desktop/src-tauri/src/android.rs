//! Android: voice goes on in the background through a foreground service
//! (gen/android: VoiceService.kt, VoicePlugin.kt), on while in voice (the
//! web client calls voice_service). Elsewhere voice_service does nothing.

use tauri::plugin::{Builder, TauriPlugin};

#[cfg(target_os = "android")]
struct Voice(tauri::plugin::PluginHandle<tauri::Wry>);

pub fn plugin() -> TauriPlugin<tauri::Wry> {
    Builder::new("voice")
        .setup(|_app, _api| {
            #[cfg(target_os = "android")]
            {
                use tauri::Manager;
                _app.manage(Voice(_api.register_android_plugin("chat.kallisti.mobile", "VoicePlugin")?));
            }
            Ok(())
        })
        .build()
}

/// The voice service on (with the channel's name for its notification) or
/// off. Async: it waits for the Kotlin side, off the main thread.
#[tauri::command]
pub async fn voice_service(app: tauri::AppHandle, on: bool, channel: Option<String>) -> Result<(), String> {
    #[cfg(target_os = "android")]
    {
        use tauri::Manager;
        let voice = app.state::<Voice>();
        let result = if on {
            voice.0.run_mobile_plugin::<serde_json::Value>("start", serde_json::json!({ "channel": channel }))
        } else {
            voice.0.run_mobile_plugin::<serde_json::Value>("stop", serde_json::json!({}))
        };
        result.map(|_| ()).map_err(|err| err.to_string())
    }
    #[cfg(not(target_os = "android"))]
    {
        let _ = (app, on, channel);
        Ok(())
    }
}
