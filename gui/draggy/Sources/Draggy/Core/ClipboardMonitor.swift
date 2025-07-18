import Foundation
import AppKit

// Core domain types
struct ClipboardFile {
    let path: String
    let name: String
    let directory: String
    
    init(path: String) {
        self.path = path
        let url = URL(fileURLWithPath: path)
        self.name = url.lastPathComponent
        self.directory = url.deletingLastPathComponent().path
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
    
    private var lastChangeCount: Int = 0
    private var eventSource: DispatchSourceTimer?
    
    func refresh() {
        let pasteboard = NSPasteboard.general
        let currentChangeCount = pasteboard.changeCount
        
        // Skip if clipboard hasn't changed
        guard currentChangeCount != lastChangeCount else { return }
        lastChangeCount = currentChangeCount
        
        let foundFiles = extractFiles(from: pasteboard)
        
        // Update and notify if changed
        if files.map(\.path) != foundFiles.map(\.path) {
            files = foundFiles
            onChange?(files)
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
        // Try multiple methods to get files
        var paths: [String] = []
        
        // Method 1: File URLs
        if let urls = pasteboard.readObjects(forClasses: [NSURL.self], options: nil) as? [URL] {
            paths = urls.compactMap { $0.isFileURL ? $0.path : nil }
        }
        
        // Method 2: File paths as strings
        if paths.isEmpty, let filePaths = pasteboard.propertyList(forType: .fileURL) as? [String] {
            paths = filePaths
        }
        
        // Method 3: Legacy type
        if paths.isEmpty, 
           let filePaths = pasteboard.propertyList(forType: NSPasteboard.PasteboardType("NSFilenamesPboardType")) as? [String] {
            paths = filePaths
        }
        
        return paths.map { ClipboardFile(path: $0) }
    }
    
    deinit {
        stopMonitoring()
    }
}