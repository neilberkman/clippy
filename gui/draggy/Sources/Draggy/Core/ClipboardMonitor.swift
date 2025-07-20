import Foundation
import AppKit
import os.log

// Core domain types
struct ClipboardFile {
    let path: String
    let name: String
    let directory: String
    let modified: Date?

    init(path: String, modified: Date? = nil) {
        self.path = path
        let url = URL(fileURLWithPath: path)
        self.name = url.lastPathComponent
        self.directory = url.deletingLastPathComponent().path
        self.modified = modified
    }
}

// Core protocol for clipboard operations
protocol ClipboardMonitor: AnyObject {
    var files: [ClipboardFile] { get }
    var onChange: (([ClipboardFile]) -> Void)? { get set }

    func refresh()
    func startMonitoring()
    func stopMonitoring()
}

// Core implementation - event-driven, no polling
class SystemClipboardMonitor: ClipboardMonitor {
    private(set) var files: [ClipboardFile] = []
    var onChange: (([ClipboardFile]) -> Void)?

    private var lastChangeCount: Int = -1  // Start with -1 to force initial load
    private var eventSource: DispatchSourceTimer?
    private let logger = Logger(subsystem: "com.neilberkman.draggy", category: "ClipboardMonitor")

    func refresh() {
        let pasteboard = NSPasteboard.general
        let currentChangeCount = pasteboard.changeCount

        // Debug output that will show in Xcode console or system logs
        NSLog("🔍 Draggy: refresh() - currentChangeCount: \(currentChangeCount), lastChangeCount: \(lastChangeCount), files.count: \(files.count)")

        // Always load on first refresh (when lastChangeCount is -1)
        let isFirstRefresh = lastChangeCount == -1

        // Skip if clipboard hasn't changed AND this isn't the first refresh
        guard currentChangeCount != lastChangeCount || isFirstRefresh else {
            NSLog("🔍 Draggy: Skipping refresh - no change detected")
            return
        }

        lastChangeCount = currentChangeCount

        let foundFiles = extractFiles(from: pasteboard)
        NSLog("🔍 Draggy: Found \(foundFiles.count) files: \(foundFiles.map { $0.path })")

        // Update and notify if changed
        if files.map(\.path) != foundFiles.map(\.path) {
            files = foundFiles
            onChange?(files)
            NSLog("🔍 Draggy: Updated files array")
        } else {
            NSLog("🔍 Draggy: Files unchanged, not updating")
        }
    }

    func startMonitoring() {
        stopMonitoring()
        refresh() // Initial check

        // Use a more efficient approach: check when app becomes active
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(appBecameActive),
            name: NSApplication.didBecomeActiveNotification,
            object: nil
        )

        // Also check when the menu is about to open (for menu bar apps)
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(menuWillOpen),
            name: NSMenu.willSendActionNotification,
            object: nil
        )
    }

    func stopMonitoring() {
        NotificationCenter.default.removeObserver(self)
        eventSource?.cancel()
        eventSource = nil
    }

    @objc private func appBecameActive() {
        refresh()
    }

    @objc private func menuWillOpen() {
        refresh()
    }

    private func extractFiles(from pasteboard: NSPasteboard) -> [ClipboardFile] {
        // Use ClippyCore to get clipboard files - proper Core/Interface separation!
        return ClippyCore.getClipboardFiles()
    }

    deinit {
        stopMonitoring()
    }
}
