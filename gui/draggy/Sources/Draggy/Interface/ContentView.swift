import SwiftUI
import UniformTypeIdentifiers


struct ContentView: View {
    @ObservedObject var viewModel: ClipboardViewModel
    @ObservedObject var updateChecker: UpdateChecker

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
                UnifiedFileListView(viewModel: viewModel)
                Divider()
                FooterView(
                    clipboardCount: viewModel.clipboardFiles.count,
                    recentCount: viewModel.recentFiles.count,
                    viewModel: viewModel
                )
            }
            .frame(width: 300, height: 400)

            if viewModel.showOnboarding {
                OnboardingView(viewModel: viewModel)
            }
        }
        .sheet(isPresented: $viewModel.showPermissionAlert) {
            PermissionView(
                folders: viewModel.permissionNeededFolders,
                isPresented: $viewModel.showPermissionAlert
            )
        }
    }
}

// MARK: - Subviews

struct HeaderView: View {
    @ObservedObject var viewModel: ClipboardViewModel

    var body: some View {
        VStack(spacing: 0) {
            // Main header row
            HStack {
                Text("Draggy")
                    .font(.headline)
                    .help("This is Draggy")

                Spacer()

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

        }
    }
}

struct UnifiedFileListView: View {
    @ObservedObject var viewModel: ClipboardViewModel
    @State private var hasCheckedOnce = false
    @State private var showFolderPopover = false

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 0) {
                // Clipboard section - super compact
                HStack {
                    Image(systemName: "clipboard")
                        .font(.subheadline)
                        .foregroundColor(.secondary)
                    Text("Clipboard")
                        .font(.subheadline)
                        .foregroundColor(.secondary)
                    if viewModel.clipboardFiles.isEmpty && hasCheckedOnce {
                        Text("•")
                            .foregroundColor(.secondary)
                            .padding(.horizontal, 4)
                        Text("Copy files to see them here")
                            .font(.caption)
                            .foregroundColor(.secondary)
                    }
                }
                .padding(.bottom, viewModel.clipboardFiles.isEmpty ? 8 : 4)

                if !viewModel.clipboardFiles.isEmpty {
                    VStack(alignment: .leading, spacing: 8) {
                        ForEach(viewModel.clipboardFiles, id: \.path) { file in
                            FileRow(file: file, onDragStarted: viewModel.onDragStarted)
                        }
                    }
                    .padding(.bottom, 8)
                }

                Divider()

                SectionHeaderView(title: "Recent Files", systemImage: "clock") {
                    if viewModel.recentDownloadsEnabled {
                        Button(action: { showFolderPopover = true }) {
                            Image(systemName: "folder")
                        }
                        .buttonStyle(.plain)
                        .help("Choose folders to search")
                        .popover(isPresented: $showFolderPopover) {
                            FolderSelectionView(viewModel: viewModel)
                        }
                    } else {
                        EmptyView()
                    }
                }

                if viewModel.recentDownloadsEnabled {
                    ActiveFoldersView(viewModel: viewModel)
                        .padding(.bottom, 2)

                    if viewModel.recentFiles.isEmpty && hasCheckedOnce {
                        EmptySectionView(
                            title: "No recent files",
                            subtitle: "Files from selected folders will appear here"
                        )
                    } else {
                        VStack(alignment: .leading, spacing: 8) {
                            ForEach(viewModel.recentFiles, id: \.path) { file in
                                FileRow(file: file, onDragStarted: viewModel.onDragStarted)
                            }
                        }
                    }
                } else {
                    DisabledRecentSectionView(viewModel: viewModel)
                }
            }
            .padding()
        }
        .onAppear {
            // Mark that we've checked at least once after a small delay
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.2) {
                hasCheckedOnce = true
            }
        }
    }
}

struct SectionHeaderView<TrailingContent: View>: View {
    let title: String
    let systemImage: String
    let trailingContent: TrailingContent

    init(title: String, systemImage: String, @ViewBuilder trailingContent: () -> TrailingContent) {
        self.title = title
        self.systemImage = systemImage
        self.trailingContent = trailingContent()
    }

    var body: some View {
        HStack {
            Image(systemName: systemImage)
                .font(.subheadline)
                .foregroundColor(.secondary)
            Text(title)
                .font(.subheadline)
                .foregroundColor(.secondary)
            Spacer()
            trailingContent
        }
    }
}

struct EmptySectionView: View {
    let title: String
    let subtitle: String

    var body: some View {
        Text(subtitle)
            .font(.caption)
            .foregroundColor(.secondary)
            .padding(.vertical, 4)
    }
}

struct DisabledRecentSectionView: View {
    @ObservedObject var viewModel: ClipboardViewModel

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("Recent files are off")
                .font(.subheadline)
                .foregroundColor(.secondary)
            Text("Enable this to show files from Downloads, Desktop, and Documents.")
                .font(.caption)
                .foregroundColor(.secondary)

            Button("Enable Recent Files") {
                viewModel.showOnboarding = true
            }
            .buttonStyle(.link)
            .padding(.top, 4)
        }
        .padding(.vertical, 6)
    }
}

// InfoBar removed - functionality moved to FooterView

struct FooterView: View {
    let clipboardCount: Int
    let recentCount: Int
    @ObservedObject var viewModel: ClipboardViewModel

    var body: some View {
        VStack(spacing: 4) {
            HStack {
                if viewModel.recentDownloadsEnabled {
                    Text("Clipboard: \(clipboardCount) • Recent: \(recentCount)")
                        .font(.caption)
                        .foregroundColor(.secondary)
                } else {
                    Text("\(clipboardCount) clipboard file\(clipboardCount == 1 ? "" : "s")")
                        .font(.caption)
                        .foregroundColor(.secondary)
                }
                
                // Show info hint about Option key on the right with better styling
                Spacer()
                
                HStack(spacing: 4) {
                    Image(systemName: "option")
                        .font(.caption)
                        .fontWeight(.medium)
                    Text("Hold to preview")
                        .font(.caption)
                        .fontWeight(.medium)
                }
                .foregroundColor(.accentColor)
                .padding(.horizontal, 8)
                .padding(.vertical, 2)
                .background(
                    RoundedRectangle(cornerRadius: 4)
                        .fill(Color.accentColor.opacity(0.1))
                )
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

                Text("Show Recent Files")
                    .font(.headline)

                Text("Draggy can show files from your Downloads, Desktop, and Documents folders in the Recent Files section.")
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
