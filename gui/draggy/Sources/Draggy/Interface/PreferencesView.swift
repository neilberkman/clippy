import SwiftUI
import LaunchAtLogin

struct PreferencesView: View {
    @AppStorage("maxFilesShown") private var maxFilesShown: Int = 20
    @AppStorage("showFullPath") private var showFullPath: Bool = false
    @AppStorage("showThumbnails") private var showThumbnails: Bool = true

    var body: some View {
        Form {
            Section("General") {
                LaunchAtLogin.Toggle("Launch at login")
            }

            Section("Display") {
                Stepper("Maximum files shown: \(maxFilesShown)",
                       value: $maxFilesShown,
                       in: 5...50,
                       step: 5)

                Toggle("Show full file paths", isOn: $showFullPath)
                    .help("Show complete paths instead of just parent directory")

                Toggle("Show file thumbnails", isOn: $showThumbnails)
                    .help("Display previews for images and PDFs instead of generic icons")
            }


            Section {
                HStack {
                    Button("Quit Draggy") {
                        NSApplication.shared.terminate(nil)
                    }
                    .foregroundColor(.red)

                    Spacer()

                    Text("v0.11.2")
                        .foregroundColor(.secondary)
                        .font(.caption)
                }
            }
        }
        .padding()
        .frame(width: 400)
    }
}
