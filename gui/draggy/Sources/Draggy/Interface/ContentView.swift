import SwiftUI
import UniformTypeIdentifiers

struct ContentView: View {
    @ObservedObject var viewModel: ClipboardViewModel
    
    var body: some View {
        VStack(spacing: 0) {
            HeaderView(viewModel: viewModel)
            Divider()
            FileListView(files: viewModel.files)
            Divider()
            FooterView(fileCount: viewModel.files.count)
        }
        .frame(width: 300, height: 400)
    }
}

// MARK: - Subviews

struct HeaderView: View {
    @ObservedObject var viewModel: ClipboardViewModel
    
    var body: some View {
        HStack {
            Text("Draggy")
                .font(.headline)
            Spacer()
            Button(action: viewModel.refresh) {
                Image(systemName: viewModel.isRefreshing ? "arrow.clockwise.circle.fill" : "arrow.clockwise")
            }
            .buttonStyle(.plain)
            .disabled(viewModel.isRefreshing)
        }
        .padding()
    }
}

struct FileListView: View {
    let files: [ClipboardFile]
    
    var body: some View {
        if files.isEmpty {
            EmptyStateView()
        } else {
            ScrollView {
                VStack(alignment: .leading, spacing: 8) {
                    ForEach(files, id: \.path) { file in
                        FileRow(file: file)
                    }
                }
                .padding()
            }
        }
    }
}

struct EmptyStateView: View {
    var body: some View {
        VStack {
            Spacer()
            Image(systemName: "doc.on.clipboard")
                .font(.largeTitle)
                .foregroundColor(.secondary)
                .padding(.bottom, 8)
            Text("No files in clipboard")
                .foregroundColor(.secondary)
            Text("Copy files with clippy first")
                .font(.caption)
                .foregroundColor(.secondary)
            Spacer()
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

struct FooterView: View {
    let fileCount: Int
    
    var body: some View {
        HStack {
            Text("\(fileCount) file\(fileCount == 1 ? "" : "s")")
                .font(.caption)
                .foregroundColor(.secondary)
            Spacer()
        }
        .padding(.horizontal)
        .padding(.vertical, 8)
    }
}

// MARK: - File Row

struct FileRow: View {
    let file: ClipboardFile
    @State private var isHovering = false
    @AppStorage("showFullPath") private var showFullPath: Bool = false
    
    private var fileURL: URL {
        URL(fileURLWithPath: file.path)
    }
    
    private var fileIcon: NSImage {
        NSWorkspace.shared.icon(forFile: file.path)
    }
    
    var body: some View {
        HStack(spacing: 12) {
            Image(nsImage: fileIcon)
                .resizable()
                .aspectRatio(contentMode: .fit)
                .frame(width: 32, height: 32)
            
            VStack(alignment: .leading, spacing: 2) {
                Text(file.name)
                    .lineLimit(1)
                    .truncationMode(.middle)
                
                Text(showFullPath ? file.path : file.directory)
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
            DragPreview(file: file, icon: fileIcon)
        }
    }
}

struct DragPreview: View {
    let file: ClipboardFile
    let icon: NSImage
    
    var body: some View {
        HStack {
            Image(nsImage: icon)
                .resizable()
                .aspectRatio(contentMode: .fit)
                .frame(width: 16, height: 16)
            Text(file.name)
                .font(.caption)
        }
        .padding(4)
        .background(Color(NSColor.controlBackgroundColor))
        .cornerRadius(4)
    }
}