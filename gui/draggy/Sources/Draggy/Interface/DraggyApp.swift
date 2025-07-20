import SwiftUI
import AppKit

// A custom NSHostingController that automatically listens for the Escape key
class EscapableHostingController<Content: View>: NSHostingController<Content> {
    var onClose: (() -> Void)?
    private var escapeMonitor: Any?

    override func viewDidAppear() {
        super.viewDidAppear()
        // Add monitor when view appears
        escapeMonitor = NSEvent.addLocalMonitorForEvents(matching: .keyDown) { [weak self] event in
            if event.keyCode == 53 { // ESC key
                self?.onClose?()
                return nil // Consume the event
            }
            return event
        }
    }

    override func viewWillDisappear() {
        super.viewWillDisappear()
        // Remove monitor when view disappears
        if let monitor = escapeMonitor {
            NSEvent.removeMonitor(monitor)
            escapeMonitor = nil
        }
    }
}

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

class AppDelegate: NSObject, NSApplicationDelegate, NSPopoverDelegate {
    private var statusItem: NSStatusItem?
    private var popover: NSPopover?
    private var eventMonitor: EventMonitor?
    private var preferencesWindow: NSWindow?
    private var updateChecker: UpdateChecker?

    // Core components
    private let clipboardMonitor: ClipboardMonitor = SystemClipboardMonitor()
    private var viewModel: ClipboardViewModel?

    func applicationDidFinishLaunching(_ notification: Notification) {
        setupStatusItem()
        setupPopover()
        setupEventMonitor()
        // Don't check for updates on launch - wait until user actually opens popover
    }

    private func setupStatusItem() {
        statusItem = NSStatusBar.system.statusItem(withLength: NSStatusItem.variableLength)

        if let button = statusItem?.button {
            button.image = NSImage(systemSymbolName: "doc.on.clipboard", accessibilityDescription: "Draggy")
            button.action = #selector(handleClick(_:))
            button.target = self
            button.sendAction(on: [.leftMouseUp, .rightMouseUp])
        }
    }

    private func setupPopover() {
        popover = NSPopover()
        popover?.contentSize = NSSize(width: 300, height: 400)
        popover?.behavior = .transient  // Default behavior
        popover?.delegate = self  // Set delegate to handle close events
        // Don't set content view controller here - we'll create fresh one each time
    }

    private func setupEventMonitor() {
        eventMonitor = EventMonitor(mask: [.leftMouseDown, .rightMouseDown]) { [weak self] event in
            guard let self = self, self.popover?.isShown ?? false else { return }

            // Check if click is outside the popover
            if let event = event {
                let popoverFrame = self.popover?.contentViewController?.view.window?.frame ?? .zero

                if !popoverFrame.contains(event.locationInWindow) {
                    self.closePopover(nil)
                }
            }
        }
    }


    @objc private func handleClick(_ sender: NSStatusBarButton) {
        let event = NSApp.currentEvent!

        if event.type == .rightMouseUp {
            showContextMenu()
        } else {
            togglePopover(sender)
        }
    }

    private func showContextMenu() {
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
    }

    @objc private func togglePopover(_ sender: AnyObject?) {
        if let popover = popover {
            if popover.isShown {
                closePopover(sender)
            } else {
                showPopover(sender)
            }
        }
    }

    private func showPopover(_ sender: AnyObject?) {
        if let button = statusItem?.button {
            // Create direct closure to control popover behavior
            let onDragStarted: () -> Void = { [weak self] in
                print("DEBUG AppDelegate: Direct closure called!")
                print("DEBUG AppDelegate: Current popover behavior: \(String(describing: self?.popover?.behavior))")
                self?.popover?.behavior = .applicationDefined
                print("DEBUG AppDelegate: Set popover behavior to .applicationDefined")
            }

            // Create fresh view model for each popover session
            viewModel = ClipboardViewModel(monitor: clipboardMonitor, onDragStarted: onDragStarted)

            // Use EscapableHostingController for ESC key support
            let hostingController = EscapableHostingController(rootView: ContentView(viewModel: viewModel!))
            hostingController.onClose = { [weak self] in
                self?.closePopover(nil)
            }
            popover?.contentViewController = hostingController

            // Refresh clipboard when showing popover
            viewModel?.refresh()
            popover?.show(relativeTo: button.bounds, of: button, preferredEdge: .minY)

            // Set window level to floating to ensure it doesn't interfere with system dialogs
            if let window = popover?.contentViewController?.view.window {
                window.level = .floating
                // Make the popover window key to receive keyboard events
                window.makeKey()
                // Make the view first responder
                window.makeFirstResponder(popover?.contentViewController?.view)
            }

            // Activate the app to ensure we can receive key events
            NSApp.activate(ignoringOtherApps: true)

            eventMonitor?.start()
        }
    }

    private func closePopover(_ sender: AnyObject?) {
        popover?.performClose(sender)
        eventMonitor?.stop()
    }

    @objc private func showPreferences() {
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

    @objc private func showAbout() {
        NSApp.orderFrontStandardAboutPanel(nil)
        NSApp.activate(ignoringOtherApps: true)
    }

    @objc private func refresh() {
        viewModel?.refresh()
    }

    // MARK: - NSPopoverDelegate

    func popoverDidClose(_ notification: Notification) {
        // Reset behavior when popover closes
        print("DEBUG AppDelegate: popoverDidClose called")
        print("DEBUG AppDelegate: Resetting behavior to .transient")
        popover?.behavior = .transient
    }

}

// MARK: - Event Monitor

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
