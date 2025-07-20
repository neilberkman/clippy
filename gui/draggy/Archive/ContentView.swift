import SwiftUI
import UniformTypeIdentifiers

struct ContentView: View {
    @ObservedObject var clipboardManager: ClipboardManager

    var body: some View {
        VStack(spacing: 0) {
            // Header
            HStack {
                Text("Draggy")
                    .font(.headline)
                Spacer()
                Button(action: clipboardManager.refresh) {
                    Image(systemName: "arrow.clockwise")
                }
                .buttonStyle(.plain)
            }
            .padding()

            Divider()

            // File list
            if clipboardManager.files.isEmpty {
                VStack {
                    Spacer()
                    Text("No files in clipboard")
                        .foregroundColor(.secondary)
                    Text("Copy files with clippy first")
                        .font(.caption)
                        .foregroundColor(.secondary)
                    Spacer()
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                ScrollView {
                    VStack(alignment: .leading, spacing: 8) {
                        ForEach(clipboardManager.files, id: \.self) { file in
                            FileRow(filePath: file)
                        }
                    }
                    .padding()
                }
            }

            Divider()

            // Footer
            HStack {
                Text("\(clipboardManager.files.count) file\(clipboardManager.files.count == 1 ? "" : "s")")
                    .font(.caption)
                    .foregroundColor(.secondary)
                Spacer()
            }
            .padding(.horizontal)
            .padding(.vertical, 8)
        }
        .frame(width: 300, height: 400)
    }
}

struct FileRow: View {
    let filePath: String
    @State private var isHovering = false

    var fileURL: URL {
        URL(fileURLWithPath: filePath)
    }

    var fileName: String {
        fileURL.lastPathComponent
    }

    var fileIcon: NSImage {
        NSWorkspace.shared.icon(forFile: filePath)
    }

    var body: some View {
        HStack(spacing: 12) {
            Image(nsImage: fileIcon)
                .resizable()
                .aspectRatio(contentMode: .fit)
                .frame(width: 32, height: 32)

            VStack(alignment: .leading, spacing: 2) {
                Text(fileName)
                    .lineLimit(1)
                    .truncationMode(.middle)

                Text(fileURL.deletingLastPathComponent().path)
                    .font(.caption)
                    .foregroundColor(.secondary)
                    .lineLimit(1)
                    .truncationMode(.middle)
            }

            Spacer()
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .background(isHovering ? Color.accentColor.opacity(0.1) : Color.clear)
        .cornerRadius(6)
        .onHover { hovering in
            isHovering = hovering
        }
        .draggable(fileURL) {
            // Drag preview
            HStack {
                Image(nsImage: fileIcon)
                    .resizable()
                    .aspectRatio(contentMode: .fit)
                    .frame(width: 16, height: 16)
                Text(fileName)
                    .font(.caption)
            }
            .padding(4)
            .background(Color(NSColor.controlBackgroundColor))
            .cornerRadius(4)
        }
    }
}

class ClipboardManager: ObservableObject {
    @Published var files: [String] = []
    @AppStorage("refreshInterval") private var refreshInterval: Double = 2.0
    @AppStorage("maxFilesShown") private var maxFilesShown: Int = 20

    var onFilesChanged: (([String]) -> Void)?
    private var timer: Timer?
    private var lastChangeCount: Int = 0

    init() {
        refresh()
        startMonitoring()
    }

    deinit {
        timer?.invalidate()
    }

    func refresh() {
        let pasteboard = NSPasteboard.general
        let currentChangeCount = pasteboard.changeCount

        // Skip if clipboard hasn't changed
        if currentChangeCount == lastChangeCount {
            return
        }
        lastChangeCount = currentChangeCount

        var foundFiles: [String] = []

        // Check for file URLs
        if let urls = pasteboard.readObjects(forClasses: [NSURL.self], options: nil) as? [URL] {
            foundFiles = urls.compactMap { url in
                // Only include file URLs (not web URLs)
                if url.isFileURL {
                    return url.path
                }
                return nil
            }
        }

        // If no URLs found, check for file paths as strings
        if foundFiles.isEmpty {
            if let filePaths = pasteboard.propertyList(forType: .fileURL) as? [String] {
                foundFiles = filePaths
            }
        }

        // Check the older NSFilenamesPboardType
        if foundFiles.isEmpty {
            if let filePaths = pasteboard.propertyList(forType: NSPasteboard.PasteboardType("NSFilenamesPboardType")) as? [String] {
                foundFiles = filePaths
            }
        }

        // Limit files shown
        if foundFiles.count > maxFilesShown {
            foundFiles = Array(foundFiles.prefix(maxFilesShown))
        }

        // Update and notify if changed
        if files != foundFiles {
            files = foundFiles
            onFilesChanged?(files)
        }
    }

    private func startMonitoring() {
        updateTimer()

        // Listen for preference changes
        NotificationCenter.default.addObserver(
            self,
            selector: #selector(updateTimer),
            name: UserDefaults.didChangeNotification,
            object: nil
        )
    }

    @objc private func updateTimer() {
        timer?.invalidate()
        timer = Timer.scheduledTimer(withTimeInterval: refreshInterval, repeats: true) { [weak self] _ in
            self?.refresh()
        }
    }
}
