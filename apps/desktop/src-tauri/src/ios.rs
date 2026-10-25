//! iOS: the web view fills the screen, under the status bar and the home
//! indicator. UIKit would inset its scroll view by the safe areas, which
//! leaves a white band below the page; the page keeps clear of them itself
//! (app.css: --safe-*).

#[cfg(target_os = "ios")]
pub fn edge_to_edge(window: &tauri::WebviewWindow) -> tauri::Result<()> {
    use objc2::{msg_send, runtime::AnyObject};
    // UIScrollViewContentInsetAdjustmentBehavior.never
    const NEVER: isize = 2;
    window.with_webview(|webview| {
        // SAFETY: on iOS the platform web view is a WKWebView (a UIView with
        // a scrollView); with_webview runs on the main thread.
        unsafe {
            let view = webview.inner().cast::<AnyObject>();
            let scroll: *mut AnyObject = msg_send![view, scrollView];
            let _: () = msg_send![scroll, setContentInsetAdjustmentBehavior: NEVER];
        }
    })
}
