import Foundation
import AppKit

// Simple update checker for Draggy that checks GitHub releases
@MainActor
class UpdateChecker: ObservableObject {
    @Published var updateAvailable = false
    @Published var latestVersion: String?
    @Published var downloadURL: URL?
    @Published var isChecking = false

    private let owner = "neilberkman"
    private let repo = "clippy"
    private let currentVersion: String

    // Store last check time to avoid excessive API calls
    private let lastCheckKey = "UpdateChecker.lastCheckTime"
    private let updateDismissedKey = "UpdateChecker.dismissedVersion"

    init() {
        // Get current version from bundle
        self.currentVersion = Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String ?? "0.0.0"
    }

    func checkForUpdatesIfNeeded() {
        // Check if we should perform the check (every 2 hours when app is actually used)
        let lastCheck = UserDefaults.standard.object(forKey: lastCheckKey) as? Date ?? Date.distantPast
        let hoursSinceLastCheck = Date().timeIntervalSince(lastCheck) / 3600

        // Only check once every 2 hours when user opens the app
        guard hoursSinceLastCheck >= 2 else { return }

        Task {
            await checkForUpdates()
        }
    }

    func checkForUpdates() async {
        guard !isChecking else { return }

        isChecking = true
        defer { isChecking = false }

        // Update last check time
        UserDefaults.standard.set(Date(), forKey: lastCheckKey)

        do {
            // GitHub API endpoint for latest release
            let url = URL(string: "https://api.github.com/repos/\(owner)/\(repo)/releases/latest")!
            
            // Create request with proper headers
            var request = URLRequest(url: url)
            request.setValue("application/vnd.github.v3+json", forHTTPHeaderField: "Accept")
            request.setValue("Draggy/\(currentVersion)", forHTTPHeaderField: "User-Agent")
            
            let (data, response) = try await URLSession.shared.data(for: request)
            
            // Check for rate limiting
            if let httpResponse = response as? HTTPURLResponse {
                if httpResponse.statusCode == 403 {
                    // Rate limited, fail silently
                    print("GitHub API rate limited")
                    return
                } else if httpResponse.statusCode >= 300 && httpResponse.statusCode < 400 {
                    // Handle redirects if needed (URLSession usually handles these automatically)
                    print("Redirect response: \(httpResponse.statusCode)")
                }
            }

            // Parse the JSON response
            if let json = try JSONSerialization.jsonObject(with: data) as? [String: Any],
               let tagName = json["tag_name"] as? String,
               let assets = json["assets"] as? [[String: Any]] {

                // Extract version from tag with more robust parsing
                let latestVersion = extractVersion(from: tagName)

                // Check if update is available
                if isNewerVersion(latestVersion, than: currentVersion) {
                    // Check if user has dismissed this version
                    let dismissedVersion = UserDefaults.standard.string(forKey: updateDismissedKey)
                    if dismissedVersion != latestVersion {
                        self.latestVersion = latestVersion

                        // Find Draggy.app.zip asset
                        if let draggyAsset = assets.first(where: { ($0["name"] as? String) == "Draggy.app.zip" }),
                           let downloadURLString = draggyAsset["browser_download_url"] as? String,
                           let downloadURL = URL(string: downloadURLString) {
                            self.downloadURL = downloadURL
                            self.updateAvailable = true
                        }
                    }
                }
            }
        } catch {
            // Silently fail - we don't want to bother users with update check errors
            print("Update check failed: \(error)")
        }
    }

    func dismissUpdate() {
        if let version = latestVersion {
            UserDefaults.standard.set(version, forKey: updateDismissedKey)
        }
        updateAvailable = false
    }

    func openDownloadPage() {
        // For Homebrew users, we should respect the package manager
        // Direct them to use brew upgrade instead of downloading manually
        if let url = downloadURL {
            NSWorkspace.shared.open(url)
        }
    }

    var updateMessage: String {
        // Check if installed via Homebrew
        if isInstalledViaHomebrew() {
            return "Run '\(getBrewUpdateCommand())' to update"
        } else {
            return "Download version \(latestVersion ?? "") from GitHub"
        }
    }

    func getBrewUpdateCommand() -> String {
        return "brew upgrade --cask draggy"
    }

    func isInstalledViaHomebrew() -> Bool {
        // Find brew in PATH using which command (respects user's shell environment)
        guard let brewPath = findBrewPath() else {
            return false
        }

        // Check if draggy is installed via brew cask
        let task = Process()
        task.executableURL = URL(fileURLWithPath: brewPath)
        task.arguments = ["list", "--cask"]

        let pipe = Pipe()
        task.standardOutput = pipe
        task.standardError = FileHandle.nullDevice

        do {
            try task.run()
            task.waitUntilExit()

            let data = pipe.fileHandleForReading.readDataToEndOfFile()
            if let output = String(data: data, encoding: .utf8) {
                // Check if draggy is in the cask list
                return output.contains("draggy")
            }
        } catch {
            // If we can't run brew, assume it's not installed via Homebrew
        }

        return false
    }

    private func findBrewPath() -> String? {
        // First, check common Homebrew installation locations directly
        // This is important because GUI apps on macOS have limited PATH
        let commonBrewPaths = [
            "/opt/homebrew/bin/brew",     // Apple Silicon Macs
            "/usr/local/bin/brew",         // Intel Macs
            "/usr/bin/brew",               // Unusual but possible
            "/home/linuxbrew/.linuxbrew/bin/brew" // Linux (just in case)
        ]
        
        for path in commonBrewPaths {
            if FileManager.default.fileExists(atPath: path) {
                return path
            }
        }
        
        // Fallback: Use /usr/bin/env which to find brew in the user's PATH
        // This might work if the app was launched from Terminal
        let task = Process()
        task.executableURL = URL(fileURLWithPath: "/usr/bin/env")
        task.arguments = ["which", "brew"]

        let pipe = Pipe()
        task.standardOutput = pipe
        task.standardError = FileHandle.nullDevice

        do {
            try task.run()
            task.waitUntilExit()

            if task.terminationStatus == 0 {
                let data = pipe.fileHandleForReading.readDataToEndOfFile()
                if let path = String(data: data, encoding: .utf8)?.trimmingCharacters(in: .whitespacesAndNewlines),
                   !path.isEmpty {
                    return path
                }
            }
        } catch {
            print("Error finding brew using which: \(error)")
        }

        return nil
    }

    // Extract version from various tag formats
    private func extractVersion(from tag: String) -> String {
        // Try different patterns:
        // - "draggy-v0.10.1" -> "0.10.1"
        // - "v0.10.1" -> "0.10.1"
        // - "0.10.1" -> "0.10.1"
        // - "draggy-0.10.1" -> "0.10.1"
        
        let patterns = [
            "draggy-v", "draggy-", "v"
        ]
        
        var version = tag
        
        // Remove known prefixes
        for pattern in patterns {
            if version.hasPrefix(pattern) {
                version = String(version.dropFirst(pattern.count))
                break
            }
        }
        
        // Validate that what remains looks like a version
        let versionComponents = version.split(separator: ".")
        if versionComponents.count >= 2 && versionComponents.allSatisfy({ Int($0) != nil }) {
            return version
        }
        
        // If validation fails, return the original tag
        return tag
    }
    
    // Simple version comparison (assumes semantic versioning)
    private func isNewerVersion(_ new: String, than current: String) -> Bool {
        let newComponents = new.split(separator: ".").compactMap { Int($0) }
        let currentComponents = current.split(separator: ".").compactMap { Int($0) }

        // Pad arrays to same length
        let maxLength = max(newComponents.count, currentComponents.count)
        let paddedNew = newComponents + Array(repeating: 0, count: maxLength - newComponents.count)
        let paddedCurrent = currentComponents + Array(repeating: 0, count: maxLength - currentComponents.count)

        // Compare component by component
        for i in 0..<maxLength {
            if paddedNew[i] > paddedCurrent[i] {
                return true
            } else if paddedNew[i] < paddedCurrent[i] {
                return false
            }
        }

        return false
    }
}
