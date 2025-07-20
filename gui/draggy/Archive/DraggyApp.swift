import SwiftUI
import AppKit

@main
struct DraggyApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) var appDelegate

    var body: some Scene {
        // No window scene - we're a menu bar app
        Settings {
            EmptyView()
        }
    }
}

class AppDelegate: NSObject, NSApplicationDelegate {
    var statusItem: NSStatusItem?
    var popover: NSPopover?
    var eventMonitor: EventMonitor?
    var preferencesWindow: NSWindow?
    let clipboardManager = ClipboardManager()

    func applicationDidFinishLaunching(_ notification: Notification) {
        // Create the status bar item
        statusItem = NSStatusBar.system.statusItem(withLength: NSStatusItem.variableLength)

        if let button = statusItem?.button {
            button.image = NSImage(systemSymbolName: "doc.on.clipboard", accessibilityDescription: "Draggy")
            button.action = #selector(handleClick(_:))
            button.target = self
            button.sendAction(on: [.leftMouseUp, .rightMouseUp])
        }

        // Create the popover
        popover = NSPopover()
        popover?.contentSize = NSSize(width: 300, height: 400)
        popover?.behavior = .transient
        popover?.contentViewController = NSHostingController(rootView: ContentView(clipboardManager: clipboardManager))

        // Monitor for clicks outside the popover
        eventMonitor = EventMonitor(mask: [.leftMouseDown, .rightMouseDown]) { [weak self] event in
            if let self = self, self.popover?.isShown ?? false {
                self.closePopover(nil)
            }
        }

        // Monitor clipboard for auto-hide feature
        clipboardManager.onFilesChanged = { [weak self] files in
            self?.updateVisibility(hasFiles: !files.isEmpty)
        }

        // Set initial visibility
        updateVisibility(hasFiles: !clipboardManager.files.isEmpty)
    }

    func updateVisibility(hasFiles: Bool) {
        let hideWhenEmpty = UserDefaults.standard.bool(forKey: "hideWhenEmpty")
        statusItem?.isVisible = !hideWhenEmpty || hasFiles
    }

    @objc func handleClick(_ sender: NSStatusBarButton) {
        let event = NSApp.currentEvent!

        if event.type == .rightMouseUp {
            // Show menu on right-click
            let menu = NSMenu()

            menu.addItem(NSMenuItem(title: "Preferences...", action: #selector(showPreferences), keyEquivalent: ","))
            menu.addItem(NSMenuItem.separator())
            menu.addItem(NSMenuItem(title: "Refresh", action: #selector(refresh), keyEquivalent: "r"))
            menu.addItem(NSMenuItem.separator())
            menu.addItem(NSMenuItem(title: "About Draggy", action: #selector(showAbout), keyEquivalent: ""))
            menu.addItem(NSMenuItem(title: "Quit Draggy", action: #selector(NSApplication.terminate(_:)), keyEquivalent: "q"))

            statusItem?.menu = menu
            statusItem?.button?.performClick(nil)
            statusItem?.menu = nil
        } else {
            // Show popover on left-click
            togglePopover(sender)
        }
    }

    @objc func showPreferences() {
        if preferencesWindow == nil {
            let window = NSWindow(
                contentRect: NSRect(x: 0, y: 0, width: 400, height: 350),
                styleMask: [.titled, .closable],
                backing: .buffered,
                defer: false
            )
            window.title = "Draggy Preferences"
            window.contentView = NSHostingView(rootView: PreferencesView())
            window.center()
            preferencesWindow = window
        }

        preferencesWindow?.makeKeyAndOrderFront(nil)
        NSApp.activate(ignoringOtherApps: true)
    }

    @objc func showAbout() {
        NSApp.orderFrontStandardAboutPanel(nil)
        NSApp.activate(ignoringOtherApps: true)
    }

    @objc func refresh() {
        clipboardManager.refresh()
    }

    @objc func togglePopover(_ sender: AnyObject?) {
        if let popover = popover {
            if popover.isShown {
                closePopover(sender)
            } else {
                showPopover(sender)
            }
        }
    }

    func showPopover(_ sender: AnyObject?) {
        if let button = statusItem?.button {
            popover?.show(relativeTo: button.bounds, of: button, preferredEdge: .minY)
            eventMonitor?.start()
        }
    }

    func closePopover(_ sender: AnyObject?) {
        popover?.performClose(sender)
        eventMonitor?.stop()
    }
}

// Event monitor for detecting clicks outside the popover
class EventMonitor {
    private var monitor: Any?
    private let mask: NSEvent.EventTypeMask
    private let handler: (NSEvent?) -> Void

    init(mask: NSEvent.EventTypeMask, handler: @escaping (NSEvent?) -> Void) {
        self.mask = mask
        self.handler = handler
    }

    deinit {
        stop()
    }

    func start() {
        monitor = NSEvent.addGlobalMonitorForEvents(matching: mask, handler: handler)
    }

    func stop() {
        if let monitor = monitor {
            NSEvent.removeMonitor(monitor)
            self.monitor = nil
        }
    }
}
