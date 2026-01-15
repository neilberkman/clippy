import SwiftUI
import Combine
import Foundation

// Interface-layer view model that bridges Core to SwiftUI
class ClipboardViewModel: ObservableObject {
    @Published var clipboardFiles: [ClipboardFile] = []
    @Published var recentFiles: [ClipboardFile] = []
    @Published var isRefreshing = false
    @Published var showOnboarding = false
    @Published var showPermissionAlert = false
    @Published var permissionNeededFolders: [String] = []

    // User preferences (interface concern)
    @AppStorage("maxFilesShown") var maxFilesShown: Int = 20
    @AppStorage("hasSeenRecentDownloadsPrompt") var hasSeenRecentDownloadsPrompt: Bool = false
    @AppStorage("recentDownloadsEnabled") var recentDownloadsEnabled: Bool = false
    @AppStorage("hasAccessedRecentDownloads") var hasAccessedRecentDownloads: Bool = false
    
    // Session-only folder overrides
    var sessionFolderOverrides: [String: Any?]?

    private let monitor: ClipboardMonitor
    let onDragStarted: (() -> Void)?

    init(monitor: ClipboardMonitor = SystemClipboardMonitor(), onDragStarted: (() -> Void)? = nil) {
        self.monitor = monitor
        self.onDragStarted = onDragStarted


        // Bridge core events to UI updates
        monitor.onChange = { [weak self] files in
            DispatchQueue.main.async {
                self?.updateFiles(files)
            }
        }

        monitor.startMonitoring()
        
        // Listen for permission errors
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(handlePermissionError),
            name: NSNotification.Name("ClippyNeedsPermission"),
            object: nil
        )

        // Immediately sync with monitor's current state
        if !monitor.files.isEmpty {
            self.updateFiles(monitor.files)
        }
        if recentDownloadsEnabled {
            loadRecentFiles()
        }

        // Check clipboard on startup with a small delay to ensure everything is initialized
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.2) {
            monitor.refresh()
        }
    }

    func refresh() {
        isRefreshing = true

        monitor.refresh()
        if recentDownloadsEnabled {
            loadRecentFiles()
        }

        // UI feedback
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) { [weak self] in
            self?.isRefreshing = false
        }
    }

    func enableRecentDownloads() {
        hasSeenRecentDownloadsPrompt = true
        recentDownloadsEnabled = true
        showOnboarding = false
        loadRecentFiles()
    }

    func declineRecentDownloads() {
        hasSeenRecentDownloadsPrompt = true
        recentDownloadsEnabled = false
        showOnboarding = false
        recentFiles = []
    }

    private func updateFiles(_ newFiles: [ClipboardFile]) {
        // Apply interface-specific constraints
        let limitedFiles = Array(newFiles.prefix(maxFilesShown))

        NSLog("🔍 Draggy: updateFiles called with \(newFiles.count) files")
        clipboardFiles = limitedFiles
        NSLog("🔍 Draggy: Set clipboardFiles array to \(limitedFiles.count) files")
    }

    func loadRecentFiles() {
        // Mark that we've accessed recent downloads
        hasAccessedRecentDownloads = true

        // Ensure our app is active and frontmost
        NSApp.activate(ignoringOtherApps: true)

        // Run on next run loop to ensure UI is ready
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }

            // Get effective folder preferences
            let folders = self.getEffectiveFolders()
            
            // Use ClippyCore to get recent downloads with folder filtering
            print("DEBUG: About to call ClippyCore.getRecentDownloads with folders: \(folders ?? ["all"])")
            self.recentFiles = ClippyCore.getRecentDownloads(maxCount: self.maxFilesShown, folders: folders)
            print("DEBUG: Got \(self.recentFiles.count) files from ClippyCore")
        }
    }
    
    private func getEffectiveFolders() -> [String]? {
        // Check if we have session overrides
        if let overrides = sessionFolderOverrides {
            var enabledFolders: [String] = []
            
            // Get AppStorage values for defaults
            let searchDownloads = UserDefaults.standard.object(forKey: "searchDownloads") as? Bool ?? true
            let searchDesktop = UserDefaults.standard.object(forKey: "searchDesktop") as? Bool ?? true
            let searchDocuments = UserDefaults.standard.object(forKey: "searchDocuments") as? Bool ?? true
            
            // Use session override or default
            let effectiveDownloads = (overrides["downloads"] as? Bool) ?? searchDownloads
            let effectiveDesktop = (overrides["desktop"] as? Bool) ?? searchDesktop
            let effectiveDocuments = (overrides["documents"] as? Bool) ?? searchDocuments
            
            if effectiveDownloads { enabledFolders.append("downloads") }
            if effectiveDesktop { enabledFolders.append("desktop") }
            if effectiveDocuments { enabledFolders.append("documents") }
            
            return enabledFolders.isEmpty ? nil : enabledFolders
        }
        
        // No session overrides, use defaults from AppStorage
        let searchDownloads = UserDefaults.standard.object(forKey: "searchDownloads") as? Bool ?? true
        let searchDesktop = UserDefaults.standard.object(forKey: "searchDesktop") as? Bool ?? true
        let searchDocuments = UserDefaults.standard.object(forKey: "searchDocuments") as? Bool ?? true
        
        var enabledFolders: [String] = []
        if searchDownloads { enabledFolders.append("downloads") }
        if searchDesktop { enabledFolders.append("desktop") }
        if searchDocuments { enabledFolders.append("documents") }
        
        // Return nil if all folders are enabled (use library defaults)
        return (enabledFolders.count == 3) ? nil : enabledFolders
    }
    
    func updateFolderSelection() {
        // Refresh recent downloads with new folder selection
        if recentDownloadsEnabled {
            loadRecentFiles()
        }
    }
    
    @objc private func handlePermissionError(_ notification: Notification) {
        if let errorMessage = notification.userInfo?["error"] as? String {
            // Extract folder paths from error message if possible
            let folders = extractFoldersFromError(errorMessage)
            
            DispatchQueue.main.async {
                self.permissionNeededFolders = folders
                self.showPermissionAlert = true
            }
        }
    }
    
    private func extractFoldersFromError(_ error: String) -> [String] {
        // Try to extract folder paths from the error message
        var folders: [String] = []
        
        if error.contains("Downloads") {
            folders.append("~/Downloads")
        }
        if error.contains("Desktop") {
            folders.append("~/Desktop")
        }
        if error.contains("Documents") {
            folders.append("~/Documents")
        }
        
        // If no specific folders found, show general folders
        if folders.isEmpty {
            folders = ["~/Downloads", "~/Desktop", "~/Documents"]
        }
        
        return folders
    }

}
