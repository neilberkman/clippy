//
//  MimeDescription.swift
//  Draggy
//
//  Provides human-friendly descriptions for MIME types
//

import Foundation

struct MimeDescription {
    // Core MIME type mappings based on common file types
    private static let mimeDescriptions: [String: String] = [
        // Documents
        "application/pdf": "PDF document",
        "application/vnd.openxmlformats-officedocument.wordprocessingml.document": "Word document",
        "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": "Excel spreadsheet",
        "application/vnd.openxmlformats-officedocument.presentationml.slideshow": "PowerPoint presentation",
        "application/msword": "Word document",
        "application/vnd.ms-excel": "Excel spreadsheet",
        "application/vnd.ms-powerpoint": "PowerPoint presentation",
        
        // Images
        "image/png": "PNG image",
        "image/jpeg": "JPEG image",
        "image/gif": "GIF image",
        "image/svg+xml": "SVG image",
        "image/webp": "WebP image",
        "image/tiff": "TIFF image",
        "image/bmp": "Bitmap image",
        "image/heic": "HEIC image",
        
        // Video
        "video/mp4": "MP4 video",
        "video/quicktime": "QuickTime video",
        "video/x-msvideo": "AVI video",
        "video/webm": "WebM video",
        
        // Audio
        "audio/mpeg": "MP3 audio",
        "audio/wav": "WAV audio",
        "audio/aac": "AAC audio",
        "audio/flac": "FLAC audio",
        
        // Archives
        "application/zip": "ZIP archive",
        "application/x-tar": "TAR archive",
        "application/x-rar-compressed": "RAR archive",
        "application/x-7z-compressed": "7-Zip archive",
        
        // Text
        "text/plain": "Text file",
        "text/html": "HTML document",
        "text/css": "CSS file",
        "text/javascript": "JavaScript file",
        "application/json": "JSON file",
        "application/xml": "XML file",
        "text/markdown": "Markdown file",
        
        // Other
        "application/octet-stream": "File"
    ]
    
    static func getDescription(for mimeType: String?) -> String {
        guard let mimeType = mimeType else { return "File" }
        
        // Check for exact match
        if let description = mimeDescriptions[mimeType] {
            return description
        }
        
        // Try to generate a reasonable description from the MIME type
        let components = mimeType.split(separator: "/")
        if components.count == 2 {
            let mainType = String(components[0])
            
            // Handle common main types
            switch mainType {
            case "image":
                return "Image"
            case "video":
                return "Video"
            case "audio":
                return "Audio"
            case "text":
                return "Text file"
            case "application":
                // Try to extract something useful from the subtype
                let subtype = String(components[1])
                if subtype.contains("zip") || subtype.contains("compressed") {
                    return "Archive"
                }
                return "Document"
            default:
                return "File"
            }
        }
        
        return "File"
    }
}