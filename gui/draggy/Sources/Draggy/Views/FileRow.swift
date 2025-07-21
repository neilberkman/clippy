//
//  FileRow.swift
//  Draggy
//
//  Created by Neil on 2025-07-20.
//

import SwiftUI
import UniformTypeIdentifiers
import AppKit
import Combine

// CRITICAL DRAG-AND-DROP SOLUTION FOR CLAUDE CODE COMPATIBILITY
//
// This custom class is THE KEY to making drag-and-drop work with Claude Code (TUI app).
// After extensive testing, we discovered that Claude Code has very specific requirements:
//
// 1. It needs BOTH file URL AND actual image data to display images properly
// 2. The type identifiers must be listed with file URL FIRST (matching CleanShot behavior)
// 3. Standard NSItemProvider with just image data or just file URL results in text URLs
//
// This solution works because:
// - It implements NSItemProviderWriting to control exactly what data is provided
// - It lists file URL first in writableTypeIdentifiersForItemProvider (critical!)
// - It provides actual image data for PNG/JPEG requests
// - It works within SwiftUI's .onDrag modifier (no need for complex AppKit overlays)
// - It uses fileURL.dataRepresentation NOT fileURL.absoluteString.data() (CRITICAL!)
//
// Without this EXACT approach, Claude Code will display "file:///..." URLs instead of [Image #X]
//
// DO NOT MODIFY without testing extensively with Claude Code drag-and-drop!
class ImageDragItem: NSObject, NSItemProviderWriting {
    let fileURL: URL
    let imageData: Data?

    init(fileURL: URL) {
        self.fileURL = fileURL
        self.imageData = try? Data(contentsOf: fileURL)
        super.init()
    }

    static var writableTypeIdentifiersForItemProvider: [String] {
        // ORDER MATTERS! File URL must be first to match CleanShot's behavior
        return [UTType.fileURL.identifier, UTType.png.identifier, UTType.jpeg.identifier]
    }

    func loadData(withTypeIdentifier typeIdentifier: String, forItemProviderCompletionHandler completionHandler: @escaping (Data?, Error?) -> Void) -> Progress? {
        if typeIdentifier == UTType.fileURL.identifier {
            // CRITICAL: Must use fileURL.dataRepresentation, NOT fileURL.absoluteString.data(using: .utf8)!
            // This is THE KEY difference that makes drag-and-drop work with Claude Code.
            // dataRepresentation provides the file URL in the exact format that macOS expects.
            // Using absoluteString.data() breaks the drag operation completely.
            // DO NOT CHANGE THIS WITHOUT EXTENSIVE TESTING WITH CLAUDE CODE!
            completionHandler(fileURL.dataRepresentation, nil)
        } else if let imageData = imageData {
            // For image types, provide the actual image data
            completionHandler(imageData, nil)
        } else {
            completionHandler(nil, nil)
        }
        return nil
    }
}

/// One row in the pop-over list.
struct FileRow: View {
    let file: ClipboardFile
    let onDragStarted: (() -> Void)?          // keeps pop-over open

    private var fileURL: URL { URL(fileURLWithPath: file.path) }

    // Preview state
    @State private var isHovering = false
    @State private var isOptionPressed = false
    @State private var showingPreview = false
    @State private var thumbnail: NSImage?
    @State private var timer: Timer?
    @State private var previewWindow: NSWindow?
    @AppStorage("showThumbnails") private var showThumbnails: Bool = true

    var body: some View {
        HStack(spacing: 12) {
            // Always show small thumbnail if available and enabled
            if showThumbnails, let thumb = thumbnail {
                Image(nsImage: thumb)
                    .resizable()
                    .aspectRatio(contentMode: .fit)
                    .frame(width: 32, height: 32)
                    .cornerRadius(4)
            } else {
                Image(nsImage: icon)
                    .resizable()
                    .frame(width: 32, height: 32)
                    .cornerRadius(4)
            }

            VStack(alignment: .leading, spacing: 2) {
                Text(fileURL.lastPathComponent).font(.subheadline)
                HStack(spacing: 4) {
                    Text(byteCount).font(.caption2).foregroundColor(.secondary)
                    Text("•").font(.caption2).foregroundColor(.secondary)
                    Text(folderSource).font(.caption2).foregroundColor(.secondary)
                }
            }
            Spacer(minLength: 0)
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 8)
        .frame(height: 54)
        .background(isHovering ? Color.accentColor.opacity(0.1) : Color.clear)
        .cornerRadius(6)
        .contentShape(Rectangle())            // full-row hit-target
        .onTapGesture(count: 2) {
            // Double-click to open file
            NSWorkspace.shared.open(fileURL)
        }
        .simultaneousGesture(
            TapGesture(count: 1)
                .onEnded { _ in
                    // Single tap does nothing but prevents drag conflicts
                }
        )
        .onHover { hovering in
            isHovering = hovering
            if hovering {
                NSCursor.pointingHand.push()
                // Start timer to check modifier state while hovering
                timer = Timer.scheduledTimer(withTimeInterval: 0.05, repeats: true) { _ in
                    let optionDown = NSEvent.modifierFlags.contains(.option)

                    if optionDown != isOptionPressed {
                        isOptionPressed = optionDown
                        if isOptionPressed && showThumbnails && thumbnail != nil {
                            showingPreview = true
                        } else {
                            showingPreview = false
                        }
                    }
                }
            } else {
                NSCursor.pop()
                timer?.invalidate()
                timer = nil
                showingPreview = false
                isOptionPressed = false
            }
        }
        .onDrag {
            onDragStarted?()

            // Check if it's an image
            let isImage = ["png", "jpg", "jpeg", "gif", "tiff", "bmp"].contains(fileURL.pathExtension.lowercased())

            if isImage {
                // Use custom drag item that provides both URL and image data
                // This is what makes Claude Code show [Image #X] instead of file:///...
                let provider = NSItemProvider(object: ImageDragItem(fileURL: fileURL))
                // Use filename without extension since Finder adds its own extension
                provider.suggestedName = fileURL.deletingPathExtension().lastPathComponent
                return provider
            } else {
                // Non-image files just provide file URL (normal behavior)
                return NSItemProvider(object: fileURL as NSURL)
            }
        }
        .onReceive(NotificationCenter.default.publisher(for: NSApplication.didBecomeActiveNotification)) { _ in
            // Check modifier keys when app becomes active
            if isHovering {
                isOptionPressed = NSEvent.modifierFlags.contains(.option)
                showingPreview = isOptionPressed && showThumbnails && thumbnail != nil
            }
        }
        .onChange(of: showingPreview) { newValue in
            if newValue, let thumbnail = thumbnail {
                showPreviewWindow(thumbnail: thumbnail)
            } else {
                hidePreviewWindow()
            }
        }
        .onAppear {
            generateThumbnail()
        }
        .onDisappear {
            timer?.invalidate()
            timer = nil
            hidePreviewWindow()
        }
    }

    // MARK: – Private helpers

    private var icon: NSImage {
        NSWorkspace.shared.icon(forFile: fileURL.path)
    }

    private func generateThumbnail() {
        guard showThumbnails else { return }

        DispatchQueue.global(qos: .background).async {
            if let thumb = ThumbnailGenerator.generateThumbnail(for: file.path, size: CGSize(width: 512, height: 512)) {
                DispatchQueue.main.async {
                    self.thumbnail = thumb
                }
            }
        }
    }

    private var byteCount: String {
        (try? FileManager.default
            .attributesOfItem(atPath: fileURL.path)[.size] as? NSNumber)
        .map {
            let f = ByteCountFormatter(); f.allowedUnits = .useAll; f.countStyle = .file
            return f.string(fromByteCount: $0.int64Value)
        } ?? ""
    }
    
    private var folderSource: String {
        let homeDir = FileManager.default.homeDirectoryForCurrentUser
        let downloadsPath = homeDir.appendingPathComponent("Downloads").path
        let desktopPath = homeDir.appendingPathComponent("Desktop").path
        let documentsPath = homeDir.appendingPathComponent("Documents").path
        
        let filePath = fileURL.path
        
        if filePath.hasPrefix(downloadsPath) {
            return "Downloads"
        } else if filePath.hasPrefix(desktopPath) {
            return "Desktop"  
        } else if filePath.hasPrefix(documentsPath) {
            return "Documents"
        } else {
            // For files not in standard folders, show the immediate parent folder name
            return fileURL.deletingLastPathComponent().lastPathComponent
        }
    }

    private func showPreviewWindow(thumbnail: NSImage) {
        // Close existing preview window if any
        hidePreviewWindow()

        // Calculate window size based on image aspect ratio
        let maxSize: CGFloat = 512
        let imageSize = thumbnail.size
        let aspectRatio = imageSize.width / imageSize.height

        var windowWidth: CGFloat
        var windowHeight: CGFloat

        if aspectRatio > 1 {
            // Landscape
            windowWidth = maxSize
            windowHeight = maxSize / aspectRatio
        } else {
            // Portrait or square
            windowHeight = maxSize
            windowWidth = maxSize * aspectRatio
        }

        // Add padding
        let padding: CGFloat = 40
        windowWidth += padding
        windowHeight += padding

        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: windowWidth, height: windowHeight),
            styleMask: [.borderless],
            backing: .buffered,
            defer: false
        )

        window.isOpaque = false
        window.backgroundColor = NSColor.clear
        window.level = .statusBar  // Same level as menu bar apps
        window.isReleasedWhenClosed = false
        window.ignoresMouseEvents = true

        // Create preview view with padding
        let containerView = NSView(frame: NSRect(x: 0, y: 0, width: windowWidth, height: windowHeight))
        containerView.wantsLayer = true

        // Use a semi-transparent background effect
        let visualEffect = NSVisualEffectView(frame: containerView.bounds)
        visualEffect.autoresizingMask = [.width, .height]
        visualEffect.material = .popover
        visualEffect.state = .active
        visualEffect.wantsLayer = true
        visualEffect.layer?.cornerRadius = 12
        visualEffect.layer?.masksToBounds = true

        containerView.layer?.cornerRadius = 12
        containerView.layer?.borderWidth = 1
        containerView.layer?.borderColor = NSColor.separatorColor.cgColor

        // Add a subtle shadow
        containerView.layer?.shadowColor = NSColor.black.cgColor
        containerView.layer?.shadowOpacity = 0.15
        containerView.layer?.shadowOffset = CGSize(width: 0, height: -2)
        containerView.layer?.shadowRadius = 10

        containerView.addSubview(visualEffect)

        let previewView = NSImageView(frame: NSRect(x: 20, y: 20, width: windowWidth - 40, height: windowHeight - 40))
        previewView.image = thumbnail
        previewView.imageScaling = .scaleProportionallyUpOrDown

        visualEffect.addSubview(previewView)
        window.contentView = containerView

        // Position window next to Draggy
        if let screen = NSScreen.main {
            var point = NSEvent.mouseLocation

            // Adjust position to be near cursor but not under it
            point.x += 20
            point.y -= windowHeight / 2

            // Keep within screen bounds
            if point.x + windowWidth > screen.frame.maxX {
                point.x = screen.frame.maxX - windowWidth - 20
            }
            if point.y < screen.frame.minY {
                point.y = screen.frame.minY
            }
            if point.y + windowHeight > screen.frame.maxY {
                point.y = screen.frame.maxY - windowHeight
            }

            window.setFrameOrigin(point)
        }

        window.orderFront(nil)
        previewWindow = window
    }

    private func hidePreviewWindow() {
        previewWindow?.close()
        previewWindow = nil
    }
}
