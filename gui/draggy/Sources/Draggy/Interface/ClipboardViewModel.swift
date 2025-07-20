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
            // Going back to clipboard
            showingRecentDownloads = false
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

            // Use ClippyCore to get recent downloads - proper Core/Interface separation!
            print("DEBUG: About to call ClippyCore.getRecentDownloads")
            self.files = ClippyCore.getRecentDownloads(maxCount: self.maxFilesShown)
            print("DEBUG: Got \(self.files.count) files from ClippyCore")

            // If no recent files, stay in recent mode but show empty state
            if self.files.isEmpty {
                self.showingRecentDownloads = true
            }
        }
    }

}
