import SwiftUI
import UniformTypeIdentifiers


struct ContentView: View {
    @ObservedObject var viewModel: ClipboardViewModel
    @StateObject private var updateChecker = UpdateChecker()

    var body: some View {
        ZStack {
            VStack(spacing: 0) {
                // Show update notification at the top if available
                if updateChecker.updateAvailable {
                    UpdateNotificationView(updateChecker: updateChecker)
                    Divider()
                }

                HeaderView(viewModel: viewModel)
                Divider()
                FileListView(files: viewModel.files, viewModel: viewModel)
                Divider()
                FooterView(fileCount: viewModel.files.count, viewModel: viewModel)
            }
            .frame(width: 300, height: 400)

            if viewModel.showOnboarding {
                OnboardingView(viewModel: viewModel)
            }
        }
        .onAppear {
            // Check for updates when view appears
            updateChecker.checkForUpdatesIfNeeded()
        }
    }
}

// MARK: - Subviews

struct HeaderView: View {
    @ObservedObject var viewModel: ClipboardViewModel
    @State private var showToggleTooltip = false
    @State private var showRefreshTooltip = false
    @State private var showFolderPopover = false

    var body: some View {
        VStack(spacing: 0) {
            // Main header row
            HStack {
                Text("Draggy")
                    .font(.headline)
                    .help("This is Draggy")

                Spacer()

                Text(viewModel.showingRecentDownloads ? "Recent Downloads" : "Clipboard")
                    .font(.subheadline)
                    .foregroundColor(.secondary)

                Spacer()

                // Folder filter button (only show in recent downloads mode)
                if viewModel.showingRecentDownloads {
                    Button(action: { showFolderPopover = true }) {
                        Image(systemName: "folder")
                    }
                    .buttonStyle(.plain)
                    .help("Choose folders to search")
                    .popover(isPresented: $showFolderPopover) {
                        FolderSelectionView(viewModel: viewModel)
                    }
                }

                // Toggle between clipboard and recent downloads
                Button(action: viewModel.toggleRecentMode) {
                    Image(systemName: viewModel.showingRecentDownloads ? "paperclip" : "clock")
                }
                .buttonStyle(.plain)
                .help(viewModel.showingRecentDownloads ? "Show clipboard" : "Show recent downloads")

                // Settings gear icon (always visible, rightmost)
                Button(action: {
                    // Send the showPreferences action to the AppDelegate
                    NSApp.sendAction(Selector(("showPreferences")), to: NSApp.delegate, from: nil)
                }) {
                    Image(systemName: "gear")
                }
                .buttonStyle(.plain)
                .help("Open Preferences")
            }
            .padding()

            // Active folders indicator (only show in recent downloads mode)
            if viewModel.showingRecentDownloads {
                ActiveFoldersView(viewModel: viewModel)
            }
        }
    }
}

struct FileListView: View {
    let files: [ClipboardFile]
    @ObservedObject var viewModel: ClipboardViewModel
    @State private var hasCheckedOnce = false

    var body: some View {
        Group {
            if files.isEmpty && hasCheckedOnce {
                EmptyStateView(showingRecentDownloads: viewModel.showingRecentDownloads, viewModel: viewModel)
            } else if !files.isEmpty {
                VStack(spacing: 0) {

                    ScrollView {
                        VStack(alignment: .leading, spacing: 8) {
                            ForEach(files, id: \.path) { file in
                                FileRow(file: file, onDragStarted: viewModel.onDragStarted)
                            }
                        }
                        .padding()
                    }
                    .safeAreaInset(edge: .bottom) {
                        InfoBar(viewModel: viewModel)
                    }
                }
            } else {
                // Loading state - show nothing initially
                Spacer()
            }
        }
        .onAppear {
            // Mark that we've checked at least once after a small delay
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.2) {
                hasCheckedOnce = true
            }
        }
    }
}

struct EmptyStateView: View {
    let showingRecentDownloads: Bool
    @ObservedObject var viewModel: ClipboardViewModel

    var body: some View {
        VStack {
            Spacer()
            Image(systemName: showingRecentDownloads ? "clock" : "clipboard")
                .font(.largeTitle)
                .foregroundColor(.secondary)
                .padding(.bottom, 8)

            if showingRecentDownloads {
                Text("No recent downloads")
                    .foregroundColor(.secondary)
                Text("Files added to Downloads folder will appear here")
                    .font(.caption)
                    .foregroundColor(.secondary)
            } else {
                Text("Clipboard is empty")
                    .foregroundColor(.secondary)
                Text("Copy files to see them here")
                    .font(.caption)
                    .foregroundColor(.secondary)

                // Show option to enable recent downloads if not already enabled
                if !viewModel.recentDownloadsEnabled && !viewModel.hasSeenRecentDownloadsPrompt {
                    Button("Show recent files when empty") {
                        viewModel.showOnboarding = true
                    }
                    .buttonStyle(.link)
                    .padding(.top, 8)
                }
            }

            Spacer()
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

struct InfoBar: View {
    @ObservedObject var viewModel: ClipboardViewModel
    @AppStorage("showInfoBar") private var showInfoBar: Bool = true

    var body: some View {
        VStack(spacing: 0) {
            // Auto-switch message
            if viewModel.showAutoSwitchMessage {
                HStack {
                    Image(systemName: "info.circle")
                        .foregroundColor(.blue)
                    Text("Nothing in clipboard, showing recent downloads")
                        .font(.caption)
                        .foregroundColor(.primary)
                    Spacer()
                    Button(action: {
                        viewModel.showAutoSwitchMessage = false
                        showInfoBar = true  // Show info bar after dismissing auto-switch message
                    }) {
                        Image(systemName: "xmark.circle.fill")
                            .foregroundColor(.secondary)
                            .font(.caption)
                    }
                    .buttonStyle(.plain)
                }
                .padding(.horizontal)
                .padding(.vertical, 8)
                .background(Color(NSColor.windowBackgroundColor).opacity(0.9))
                Divider()
            }

            // Regular info bar - only show if auto-switch message is not showing
            if showInfoBar && !viewModel.showAutoSwitchMessage {
                HStack {
                    Image(systemName: "info.circle")
                        .foregroundColor(.secondary)
                    Text("Hold ⌥ Option while hovering to preview • Drag files to other apps")
                        .font(.caption)
                        .foregroundColor(.secondary)
                    Spacer()
                    Button(action: { showInfoBar = false }) {
                        Image(systemName: "xmark.circle.fill")
                            .foregroundColor(.secondary)
                            .font(.caption)
                    }
                    .buttonStyle(.plain)
                }
                .padding(.horizontal)
                .padding(.vertical, 8)
                .background(Color(NSColor.windowBackgroundColor).opacity(0.9))
            }
        }
    }
}

struct FooterView: View {
    let fileCount: Int
    @ObservedObject var viewModel: ClipboardViewModel

    var body: some View {
        VStack(spacing: 4) {
            // Show opt-in reminder if user declined but hasn't re-enabled
            if viewModel.hasSeenRecentDownloadsPrompt && !viewModel.recentDownloadsEnabled && !viewModel.showingRecentDownloads {
                Button(action: { viewModel.showOnboarding = true }) {
                    HStack {
                        Image(systemName: "clock.arrow.circlepath")
                            .font(.caption)
                        Text("Show recent files when clipboard empty")
                            .font(.caption)
                    }
                }
                .buttonStyle(.plain)
                .foregroundColor(.accentColor)
                .padding(.vertical, 4)
            }

            HStack {
                Text("\(fileCount) file\(fileCount == 1 ? "" : "s")")
                    .font(.caption)
                    .foregroundColor(.secondary)
                Spacer()
            }
        }
        .padding(.horizontal)
        .padding(.vertical, 8)
    }
}

// MARK: - File Row

// FileRow moved to its own file


// MARK: - Onboarding

struct OnboardingView: View {
    @ObservedObject var viewModel: ClipboardViewModel
    @State private var showRelaunchWarning = false

    var body: some View {
        ZStack {
            VStack(spacing: 16) {
                Image(systemName: "clock.arrow.circlepath")
                    .font(.system(size: 40))
                    .foregroundColor(.accentColor)

                Text("Show Recent Files When Empty")
                    .font(.headline)

                Text("Draggy can automatically show files from your Downloads, Desktop, and Documents folders when the clipboard is empty.")
                    .multilineTextAlignment(.center)
                    .foregroundColor(.secondary)
                    .font(.caption)
                    .padding(.horizontal, 8)
                    .fixedSize(horizontal: false, vertical: true)

                if !showRelaunchWarning {
                    Text("This requires permission to access your Downloads, Desktop, and Documents folders.")
                        .font(.caption2)
                        .multilineTextAlignment(.center)
                        .foregroundColor(.secondary)
                        .padding(.horizontal, 8)

                    HStack(spacing: 16) {
                        Button("Not now") {
                            viewModel.declineRecentDownloads()
                        }
                        .buttonStyle(.plain)
                        .foregroundColor(.secondary)

                        Button("Enable") {
                            showRelaunchWarning = true
                            // Don't auto-trigger permissions
                        }
                        .buttonStyle(.borderedProminent)
                    }
                } else {
                    VStack(spacing: 8) {
                        Text("macOS will now ask for permission to access your Downloads, Desktop, and Documents folders")
                            .font(.caption)
                            .foregroundColor(.primary)
                            .multilineTextAlignment(.center)

                        Text("You'll need to reopen Draggy after granting these permissions")
                            .font(.caption2)
                            .foregroundColor(.secondary)
                            .multilineTextAlignment(.center)
                    }
                    .padding(12)
                    .background(Color(NSColor.controlBackgroundColor).opacity(0.5))
                    .cornerRadius(8)
                    .overlay(
                        RoundedRectangle(cornerRadius: 8)
                            .stroke(Color(NSColor.separatorColor), lineWidth: 1)
                    )

                    Spacer()
                        .frame(height: 8)

                    Button("Continue →") {
                        viewModel.enableRecentDownloads()
                    }
                    .buttonStyle(.borderedProminent)
                    .controlSize(.small)
                }
            }
            .padding(24)
            .background(Color(NSColor.windowBackgroundColor))
            .cornerRadius(10)
            .shadow(radius: 10)
            .frame(width: 300, height: 400)
            .background(Color.black.opacity(0.3))
        }
    }
}

// MARK: - Folder Selection

struct FolderSelectionView: View {
    @ObservedObject var viewModel: ClipboardViewModel
    @AppStorage("searchDownloads") private var searchDownloads: Bool = true
    @AppStorage("searchDesktop") private var searchDesktop: Bool = true
    @AppStorage("searchDocuments") private var searchDocuments: Bool = true
    
    // Session-only overrides
    @State private var sessionDownloads: Bool?
    @State private var sessionDesktop: Bool?
    @State private var sessionDocuments: Bool?
    
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Search Folders")
                .font(.headline)
            
            Text("Choose which folders to search in this session")
                .font(.caption)
                .foregroundColor(.secondary)
            
            VStack(alignment: .leading, spacing: 8) {
                Toggle("Downloads", isOn: Binding(
                    get: { sessionDownloads ?? searchDownloads },
                    set: { 
                        sessionDownloads = $0
                        viewModel.sessionFolderOverrides = [
                            "downloads": sessionDownloads,
                            "desktop": sessionDesktop,
                            "documents": sessionDocuments
                        ]
                        viewModel.updateFolderSelection() 
                    }
                ))
                
                Toggle("Desktop", isOn: Binding(
                    get: { sessionDesktop ?? searchDesktop },
                    set: { 
                        sessionDesktop = $0
                        viewModel.sessionFolderOverrides = [
                            "downloads": sessionDownloads,
                            "desktop": sessionDesktop,
                            "documents": sessionDocuments
                        ]
                        viewModel.updateFolderSelection() 
                    }
                ))
                
                Toggle("Documents", isOn: Binding(
                    get: { sessionDocuments ?? searchDocuments },
                    set: { 
                        sessionDocuments = $0
                        viewModel.sessionFolderOverrides = [
                            "downloads": sessionDownloads,
                            "desktop": sessionDesktop,
                            "documents": sessionDocuments
                        ]
                        viewModel.updateFolderSelection() 
                    }
                ))
            }
            
            Divider()
            
            HStack {
                Text("Session only • Change defaults in Preferences")
                    .font(.caption2)
                    .foregroundColor(.secondary)
                
                Spacer()
            }
        }
        .padding()
        .frame(width: 200)
        .onAppear {
            // Initialize session state from existing overrides if available
            if let existing = viewModel.sessionFolderOverrides {
                sessionDownloads = existing["downloads"] as? Bool
                sessionDesktop = existing["desktop"] as? Bool
                sessionDocuments = existing["documents"] as? Bool
            }
            // Store current session state in viewModel
            viewModel.sessionFolderOverrides = [
                "downloads": sessionDownloads,
                "desktop": sessionDesktop,
                "documents": sessionDocuments
            ]
        }
    }
}

struct ActiveFoldersView: View {
    @ObservedObject var viewModel: ClipboardViewModel
    @AppStorage("searchDownloads") private var searchDownloads: Bool = true
    @AppStorage("searchDesktop") private var searchDesktop: Bool = true
    @AppStorage("searchDocuments") private var searchDocuments: Bool = true
    
    private var activeFolders: [String] {
        var folders: [String] = []
        
        let effectiveDownloads = viewModel.sessionFolderOverrides?["downloads"] as? Bool ?? searchDownloads
        let effectiveDesktop = viewModel.sessionFolderOverrides?["desktop"] as? Bool ?? searchDesktop
        let effectiveDocuments = viewModel.sessionFolderOverrides?["documents"] as? Bool ?? searchDocuments
        
        if effectiveDownloads { folders.append("Downloads") }
        if effectiveDesktop { folders.append("Desktop") }
        if effectiveDocuments { folders.append("Documents") }
        
        return folders
    }
    
    var body: some View {
        // Only show if not searching all folders (don't show redundant info)
        if activeFolders.count < 3 && !activeFolders.isEmpty {
            HStack(spacing: 4) {
                Image(systemName: "folder")
                    .font(.caption2)
                    .foregroundColor(.secondary)
                
                Text(activeFolders.joined(separator: " • "))
                    .font(.caption2)
                    .foregroundColor(.secondary)
                
                Spacer()
            }
            .padding(.horizontal)
            .padding(.bottom, 2)
        }
    }
}
