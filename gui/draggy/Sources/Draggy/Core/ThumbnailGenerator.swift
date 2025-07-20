import Foundation
import AppKit
import Quartz
import QuickLookThumbnailing
import UniformTypeIdentifiers

class ThumbnailGenerator {
    static let shared = ThumbnailGenerator()
    private let thumbnailCache = NSCache<NSString, NSImage>()

    private init() {
        thumbnailCache.countLimit = 100 // Limit cache size
    }

    static func generateThumbnail(for path: String, size: CGSize) -> NSImage? {
        let cacheKey = "\(path)-\(Int(size.width))x\(Int(size.height))" as NSString

        // Check cache first
        if let cached = shared.thumbnailCache.object(forKey: cacheKey) {
            return cached
        }

        let fileURL = URL(fileURLWithPath: path)

        // Try different methods based on file type
        if let thumbnail = generateQuickLookThumbnail(for: fileURL, size: size) {
            shared.thumbnailCache.setObject(thumbnail, forKey: cacheKey)
            return thumbnail
        } else if let thumbnail = generateImageThumbnail(for: fileURL, size: size) {
            shared.thumbnailCache.setObject(thumbnail, forKey: cacheKey)
            return thumbnail
        } else if let thumbnail = generatePDFThumbnail(for: fileURL, size: size) {
            shared.thumbnailCache.setObject(thumbnail, forKey: cacheKey)
            return thumbnail
        }

        return nil
    }

    // Modern QuickLook thumbnail generation (macOS 10.15+)
    private static func generateQuickLookThumbnail(for url: URL, size: CGSize) -> NSImage? {
        if #available(macOS 10.15, *) {
            let scale = NSScreen.main?.backingScaleFactor ?? 2.0
            let request = QLThumbnailGenerator.Request(
                fileAt: url,
                size: size,
                scale: scale,
                representationTypes: .thumbnail
            )

            var thumbnail: NSImage?
            let semaphore = DispatchSemaphore(value: 0)

            QLThumbnailGenerator.shared.generateRepresentations(for: request) { (representation, type, error) in
                defer { semaphore.signal() }

                if let representation = representation {
                    thumbnail = representation.nsImage
                }
            }

            // Wait for completion (with timeout)
            _ = semaphore.wait(timeout: .now() + 2.0)
            return thumbnail
        }

        return nil
    }

    // Direct image thumbnail generation
    private static func generateImageThumbnail(for url: URL, size: CGSize) -> NSImage? {
        guard let uti = UTType(filenameExtension: url.pathExtension),
              uti.conforms(to: .image) else { return nil }

        guard let image = NSImage(contentsOf: url) else { return nil }

        return resizeImage(image, to: size)
    }

    // PDF thumbnail generation
    private static func generatePDFThumbnail(for url: URL, size: CGSize) -> NSImage? {
        guard url.pathExtension.lowercased() == "pdf" else { return nil }

        guard let pdfDocument = PDFDocument(url: url),
              let firstPage = pdfDocument.page(at: 0) else { return nil }

        let pageBounds = firstPage.bounds(for: .mediaBox)
        let scale = min(size.width / pageBounds.width, size.height / pageBounds.height)
        let scaledSize = CGSize(width: pageBounds.width * scale, height: pageBounds.height * scale)

        let image = NSImage(size: scaledSize)
        image.lockFocus()

        if let context = NSGraphicsContext.current?.cgContext {
            context.setFillColor(NSColor.white.cgColor)
            context.fill(CGRect(origin: .zero, size: scaledSize))

            context.scaleBy(x: scale, y: scale)
            firstPage.draw(with: .mediaBox, to: context)
        }

        image.unlockFocus()
        return image
    }

    // Helper to resize images maintaining aspect ratio
    private static func resizeImage(_ image: NSImage, to targetSize: CGSize) -> NSImage? {
        let imageSize = image.size
        let widthRatio = targetSize.width / imageSize.width
        let heightRatio = targetSize.height / imageSize.height
        let scale = min(widthRatio, heightRatio)

        let scaledSize = CGSize(
            width: imageSize.width * scale,
            height: imageSize.height * scale
        )

        let newImage = NSImage(size: scaledSize)
        newImage.lockFocus()

        NSGraphicsContext.current?.imageInterpolation = .high
        image.draw(in: NSRect(origin: .zero, size: scaledSize),
                   from: NSRect(origin: .zero, size: imageSize),
                   operation: .copy,
                   fraction: 1.0)

        newImage.unlockFocus()
        return newImage
    }
}

// Extension to convert QLThumbnailRepresentation to NSImage
@available(macOS 10.15, *)
extension QLThumbnailRepresentation {
    var nsImage: NSImage? {
        return NSImage(cgImage: cgImage, size: NSSize(width: cgImage.width, height: cgImage.height))
    }
}
