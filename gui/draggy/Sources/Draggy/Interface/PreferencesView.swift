import SwiftUI
import LaunchAtLogin

struct PreferencesView: View {
    @AppStorage("maxFilesShown") private var maxFilesShown: Int = 20
    @AppStorage("showFullPath") private var showFullPath: Bool = false
    @AppStorage("showThumbnails") private var showThumbnails: Bool = true
    @AppStorage("searchDownloads") private var searchDownloads: Bool = true
    @AppStorage("searchDesktop") private var searchDesktop: Bool = true
    @AppStorage("searchDocuments") private var searchDocuments: Bool = true
    
    let onDone: (() -> Void)?
    
    init(onDone: (() -> Void)? = nil) {
        self.onDone = onDone
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 20) {
            // General Section
            VStack(alignment: .leading, spacing: 12) {
                Text("General")
                    .font(.headline)
                
                VStack(alignment: .leading, spacing: 8) {
                    LaunchAtLogin.Toggle("Launch at login")
                }
                .padding(.leading, 16)
            }
            
            // Search Folders Section  
            VStack(alignment: .leading, spacing: 12) {
                Text("Search Folders")
                    .font(.headline)
                
                VStack(alignment: .leading, spacing: 8) {
                    Toggle("Search Downloads folder", isOn: $searchDownloads)
                        .help("Include files from ~/Downloads")
                    
                    Toggle("Search Desktop folder", isOn: $searchDesktop)
                        .help("Include files from ~/Desktop")
                    
                    Toggle("Search Documents folder", isOn: $searchDocuments)
                        .help("Include files from ~/Documents")
                }
                .padding(.leading, 16)
                
                Text("Choose which folders to search for recent files")
                    .font(.caption)
                    .foregroundColor(.secondary)
                    .padding(.leading, 16)
            }

            // Display Section
            VStack(alignment: .leading, spacing: 12) {
                Text("Display")
                    .font(.headline)
                
                VStack(alignment: .leading, spacing: 8) {
                    Stepper("Maximum files shown: \(maxFilesShown)",
                           value: $maxFilesShown,
                           in: 5...50,
                           step: 5)

                    Toggle("Show full file paths", isOn: $showFullPath)
                        .help("Show complete paths instead of just parent directory")

                    Toggle("Show file thumbnails", isOn: $showThumbnails)
                        .help("Display previews for images and PDFs instead of generic icons")
                }
                .padding(.leading, 16)
            }
            
            Spacer()

            // Bottom section
            HStack {
                Button("Quit Draggy") {
                    NSApplication.shared.terminate(nil)
                }
                .foregroundColor(.red)

                Spacer()

                Text("v0.14.0")
                    .foregroundColor(.secondary)
                    .font(.caption)
                
                if let onDone = onDone {
                    Button("Done") {
                        print("DEBUG: Done button clicked")
                        onDone()
                        print("DEBUG: onDone callback completed")
                    }
                    .buttonStyle(.borderedProminent)
                }
            }
        }
        .padding(20)
        .frame(width: 400, height: 450)
    }
}
