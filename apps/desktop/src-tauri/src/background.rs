//! Calls go on while the app is in the background. On macOS the window's
//! web view keeps running when hidden (tauri.conf.json: backgroundThrottling
//! "disabled", WKPreferences.inactiveSchedulingPolicy none, macOS 14+), and
//! here: no App Nap, which would slow the whole app down. Otherwise the page
//! stops answering the server's pings and loses the connection.

/// Turns App Nap off for the rest of the process's life (the idle system
/// may still sleep).
#[cfg(target_os = "macos")]
pub fn no_app_nap() {
    use objc2_foundation::{ns_string, NSActivityOptions, NSProcessInfo};
    let activity = NSProcessInfo::processInfo()
        .beginActivityWithOptions_reason(NSActivityOptions::UserInitiatedAllowingIdleSystemSleep, ns_string!("Voice chat"));
    std::mem::forget(activity); // ends when released
}

#[cfg(test)]
mod tests {
    #[cfg(target_os = "macos")]
    #[test]
    fn app_nap_off() {
        super::no_app_nap();
    }

    // iOS: WebKit mutes a hidden page's microphone unless the app has the
    // audio background mode.
    #[test]
    fn ios_microphone_not_muted_in_background() {
        let plist = include_str!("../Info.ios.plist");
        let modes = plist.split("<key>UIBackgroundModes</key>").nth(1).expect("UIBackgroundModes");
        let modes = &modes[..modes.find("</array>").expect("an array")];
        assert!(modes.contains("<string>audio</string>"));
        assert!(plist.contains("<key>NSMicrophoneUsageDescription</key>"));
    }

    // Export compliance is answered per build in App Store Connect: the app
    // has standard algorithms of its own (the Mumble bridge: TLS with
    // rustls, AES-OCB2), and a "true" here makes uploads need the code of
    // encryption documentation, which there is none of yet.
    #[test]
    fn ios_leaves_export_compliance_to_app_store_connect() {
        let plist = include_str!("../Info.ios.plist");
        assert!(!plist.contains("ITSAppUsesNonExemptEncryption"));
    }

    #[test]
    fn web_view_not_suspended_in_background() {
        let conf: serde_json::Value = serde_json::from_str(include_str!("../tauri.conf.json")).unwrap();
        assert_eq!(conf["app"]["windows"][0]["backgroundThrottling"], "disabled");
    }
}
