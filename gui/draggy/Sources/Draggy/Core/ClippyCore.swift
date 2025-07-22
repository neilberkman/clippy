import Foundation

// C function declarations from libclippy
@_silgen_name("ClippyGetRecentDownloads")
func ClippyGetRecentDownloads(_ maxCount: CInt, _ durationSecs: CInt, _ outError: UnsafeMutablePointer<UnsafeMutablePointer<CChar>?>) -> UnsafeMutablePointer<UnsafeMutablePointer<CChar>?>?

@_silgen_name("ClippyGetRecentDownloadsWithFolders")
func ClippyGetRecentDownloadsWithFolders(_ maxCount: CInt, _ durationSecs: CInt, _ folders: UnsafeMutablePointer<CChar>, _ outError: UnsafeMutablePointer<UnsafeMutablePointer<CChar>?>) -> UnsafeMutablePointer<UnsafeMutablePointer<CChar>?>?

@_silgen_name("ClippyFreeStringArray")
func ClippyFreeStringArray(_ arr: UnsafeMutablePointer<UnsafeMutablePointer<CChar>?>?)

@_silgen_name("ClippyCopyFile")
func ClippyCopyFile(_ path: UnsafeMutablePointer<CChar>, _ outError: UnsafeMutablePointer<UnsafeMutablePointer<CChar>?>) -> CInt

@_silgen_name("ClippyCopyText")
func ClippyCopyText(_ text: UnsafeMutablePointer<CChar>, _ outError: UnsafeMutablePointer<UnsafeMutablePointer<CChar>?>) -> CInt

@_silgen_name("ClippyGetClipboardFiles")
func ClippyGetClipboardFiles(_ outError: UnsafeMutablePointer<UnsafeMutablePointer<CChar>?>) -> UnsafeMutablePointer<UnsafeMutablePointer<CChar>?>?

// Swift wrapper for the Clippy C library
// This is the ONLY place where we interface with the Go core library
struct ClippyCore {

    // MARK: - Recent Downloads

    static func getRecentDownloads(maxCount: Int = 10, maxAge: TimeInterval? = nil, folders: [String]? = nil) -> [ClipboardFile] {
        // Use library default if no maxAge specified (0 means use default)
        let durationSecs = CInt(maxAge ?? 0)
        var errorPtr: UnsafeMutablePointer<CChar>? = nil

        // If specific folders are provided, use the folder-specific function
        if let folders = folders {
            let folderString = folders.joined(separator: ",")
            NSLog("DEBUG: Calling ClippyGetRecentDownloadsWithFolders with maxCount=\(maxCount), duration=\(durationSecs), folders=\(folderString)")
            
            guard let cStrings = folderString.withCString({ cFolders in
                ClippyGetRecentDownloadsWithFolders(CInt(maxCount), durationSecs, UnsafeMutablePointer(mutating: cFolders), &errorPtr)
            }) else {
                if let error = errorPtr {
                    let errorMessage = String(cString: error)
                    NSLog("DEBUG: ClippyGetRecentDownloadsWithFolders returned error: %@", errorMessage)
                    free(error)
                }
                return []
            }
            
            return parseFileResults(cStrings: cStrings)
        } else {
            NSLog("DEBUG: Calling ClippyGetRecentDownloads with maxCount=\(maxCount), duration=\(durationSecs)")

            guard let cStrings = ClippyGetRecentDownloads(CInt(maxCount), durationSecs, &errorPtr) else {
                if let error = errorPtr {
                    let errorMessage = String(cString: error)
                    NSLog("DEBUG: ClippyGetRecentDownloads returned error: %@", errorMessage)
                    free(error)
                } else {
                    NSLog("DEBUG: ClippyGetRecentDownloads returned nil (no files found)")
                }
                return []
            }

            return parseFileResults(cStrings: cStrings)
        }
    }
    
    private static func parseFileResults(cStrings: UnsafeMutablePointer<UnsafeMutablePointer<CChar>?>?) -> [ClipboardFile] {
        guard let cStrings = cStrings else { return [] }
        
        NSLog("DEBUG: ClippyGetRecentDownloads returned non-nil")

        // Convert C string array to Swift
        var files: [ClipboardFile] = []
        var ptr = cStrings

        while let cStr = ptr.pointee {
            let str = String(cString: cStr)

            // Parse format: path|name|unix_timestamp|mime_type
            let parts = str.split(separator: "|")
            if parts.count >= 3 {
                let path = String(parts[0])
                let timestamp = TimeInterval(parts[2]) ?? 0
                let modified = Date(timeIntervalSince1970: timestamp)
                let mimeType = parts.count >= 4 ? String(parts[3]) : nil

                files.append(ClipboardFile(path: path, modified: modified, mimeType: mimeType))
            }

            ptr = ptr.advanced(by: 1)
        }

        // Free the C memory
        ClippyFreeStringArray(cStrings)

        NSLog("DEBUG: Successfully received %d files from Go", files.count)
        return files
    }

    // MARK: - Clipboard Operations

    static func getClipboardFiles() -> [ClipboardFile] {
        var errorPtr: UnsafeMutablePointer<CChar>? = nil

        print("DEBUG: Calling ClippyGetClipboardFiles")

        guard let cStrings = ClippyGetClipboardFiles(&errorPtr) else {
            if let error = errorPtr {
                let errorMessage = String(cString: error)
                print("DEBUG: ClippyGetClipboardFiles returned error: \(errorMessage)")
                free(error)
            } else {
                print("DEBUG: ClippyGetClipboardFiles returned nil (no files)")
            }
            return []
        }

        print("DEBUG: ClippyGetClipboardFiles returned non-nil")

        var files: [ClipboardFile] = []
        var ptr = cStrings

        while let cStr = ptr.pointee {
            let path = String(cString: cStr)
            print("DEBUG: Got clipboard file: \(path)")
            files.append(ClipboardFile(path: path))
            ptr = ptr.advanced(by: 1)
        }

        ClippyFreeStringArray(cStrings)

        print("DEBUG: Successfully received \(files.count) clipboard files from Go")
        return files
    }

    static func copyFile(_ path: String) -> Bool {
        var errorPtr: UnsafeMutablePointer<CChar>? = nil

        let result = path.withCString { cPath in
            ClippyCopyFile(UnsafeMutablePointer(mutating: cPath), &errorPtr) == 1
        }

        if let error = errorPtr {
            let errorMessage = String(cString: error)
            print("DEBUG: ClippyCopyFile error: \(errorMessage)")
            free(error)
        }

        return result
    }

    static func copyText(_ text: String) -> Bool {
        var errorPtr: UnsafeMutablePointer<CChar>? = nil

        let result = text.withCString { cText in
            ClippyCopyText(UnsafeMutablePointer(mutating: cText), &errorPtr) == 1
        }

        if let error = errorPtr {
            let errorMessage = String(cString: error)
            print("DEBUG: ClippyCopyText error: \(errorMessage)")
            free(error)
        }

        return result
    }
}
