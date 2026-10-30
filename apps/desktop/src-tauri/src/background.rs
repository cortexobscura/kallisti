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

    // The iOS app has no Mumble bridge (see lib.rs): only Apple's
    // networking in WebKit, exempt encryption. Said in the plist, App Store
    // Connect does not ask per build.
    #[test]
    fn ios_declares_no_non_exempt_encryption() {
        let plist = include_str!("../Info.ios.plist");
        let value = plist.split("<key>ITSAppUsesNonExemptEncryption</key>").nth(1).expect("the key");
        assert!(value.trim_start().starts_with("<false/>"));
    }

    // The bridge's crates (TLS, AES-OCB2) are not in the iOS build.
    #[test]
    fn ios_build_has_no_bridge_crypto() {
        let cargo = include_str!("../Cargo.toml");
        let section = |header: &str| {
            let rest = cargo.split(header).nth(1).expect(header);
            rest[..rest.find("\n[").unwrap_or(rest.len())].to_string()
        };
        let everywhere = section("[dependencies]");
        let not_ios = section("[target.'cfg(not(target_os = \"ios\"))'.dependencies]");
        for c in ["rustls", "tokio-rustls", "aes", "tokio-tungstenite", "webpki-roots"] {
            assert!(!everywhere.contains(&format!("\n{c} ")), "{c} in every build");
            assert!(not_ios.contains(&format!("\n{c} ")), "{c} not outside iOS");
        }
    }

    #[test]
    fn web_view_not_suspended_in_background() {
        let conf: serde_json::Value = serde_json::from_str(include_str!("../tauri.conf.json")).unwrap();
        assert_eq!(conf["app"]["windows"][0]["backgroundThrottling"], "disabled");
    }
}
