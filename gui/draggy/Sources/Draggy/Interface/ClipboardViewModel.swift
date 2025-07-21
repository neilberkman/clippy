import SwiftUI
import Combine
import Foundation

// Interface-layer view model that bridges Core to SwiftUI
class ClipboardViewModel: ObservableObject {
    @Published var files: [ClipboardFile] = []
    @Published var isRefreshing = false
    @Published var showingRecentDownloads = false
    @Published var showOnboarding = false
    @Published var showAutoSwitchMessage = false

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

        // Immediately sync with monitor's current state
        if !monitor.files.isEmpty {
            self.updateFiles(monitor.files)
        } else if recentDownloadsEnabled {
            // If clipboard is empty on startup and recent downloads is enabled, auto-switch
            autoSwitchToRecentIfNeeded()
        }

        // Check clipboard on startup with a small delay to ensure everything is initialized
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.2) {
            // Only refresh if we haven't already auto-switched
            if !self.showingRecentDownloads {
                monitor.refresh()
            }
        }
    }

    func refresh() {
        isRefreshing = true

        if showingRecentDownloads {
            // Refresh recent downloads
            loadRecentDownloads()
        } else {
            // Refresh clipboard
            monitor.refresh()
        }

        // UI feedback
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) { [weak self] in
            self?.isRefreshing = false
        }
    }

    func toggleRecentMode() {
        // Clear auto-switch message when manually toggling
        showAutoSwitchMessage = false

        if showingRecentDownloads {
            // Going back to clipboard - clear session folder overrides
            showingRecentDownloads = false
            sessionFolderOverrides = nil
            monitor.refresh()
            // Force show clipboard files immediately
            files = Array(monitor.files.prefix(maxFilesShown))
        } else {
            // ALWAYS show onboarding if not enabled yet
            if !recentDownloadsEnabled {
                showOnboarding = true
            } else {
                // Only switch if already enabled
                showingRecentDownloads = true
                loadRecentDownloads()
            }
        }
    }

    func enableRecentDownloads() {
        hasSeenRecentDownloadsPrompt = true
        recentDownloadsEnabled = true
        showOnboarding = false
        showingRecentDownloads = true
        // Load immediately - don't close popover
        loadRecentDownloads()
    }

    func declineRecentDownloads() {
        hasSeenRecentDownloadsPrompt = true
        recentDownloadsEnabled = false
        showOnboarding = false
    }

    private func updateFiles(_ newFiles: [ClipboardFile]) {
        // Apply interface-specific constraints
        let limitedFiles = Array(newFiles.prefix(maxFilesShown))

        NSLog("🔍 Draggy: updateFiles called with \(newFiles.count) files, showingRecentDownloads=\(showingRecentDownloads)")

        // If we're showing clipboard mode
        if !showingRecentDownloads {
            files = limitedFiles
            NSLog("🔍 Draggy: Set files array to \(limitedFiles.count) files")

            // Auto-switch to recent downloads if clipboard is empty and recent downloads is enabled
            if limitedFiles.isEmpty && recentDownloadsEnabled {
                autoSwitchToRecentIfNeeded()
            }
        }
    }

    private func autoSwitchToRecentIfNeeded() {
        NSLog("🔍 Draggy: Auto-switching to recent downloads")
        showingRecentDownloads = true
        showAutoSwitchMessage = true
        // Clear any previous session folder overrides when auto-switching
        sessionFolderOverrides = nil
        DispatchQueue.main.async {
            self.loadRecentDownloads()
        }
    }

    func loadRecentDownloads() {
        // Mark that we've accessed recent downloads
        hasAccessedRecentDownloads = true
        showingRecentDownloads = true  // Ensure we stay in recent mode

        // Ensure our app is active and frontmost
        NSApp.activate(ignoringOtherApps: true)

        // Run on next run loop to ensure UI is ready
        DispatchQueue.main.async { [weak self] in
            guard let self = self else { return }

            // Get effective folder preferences
            let folders = self.getEffectiveFolders()
            
            // Use ClippyCore to get recent downloads with folder filtering
            print("DEBUG: About to call ClippyCore.getRecentDownloads with folders: \(folders ?? ["all"])")
            self.files = ClippyCore.getRecentDownloads(maxCount: self.maxFilesShown, folders: folders)
            print("DEBUG: Got \(self.files.count) files from ClippyCore")

            // If no recent files, stay in recent mode but show empty state
            if self.files.isEmpty {
                self.showingRecentDownloads = true
            }
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
        if showingRecentDownloads {
            loadRecentDownloads()
        }
    }

}
