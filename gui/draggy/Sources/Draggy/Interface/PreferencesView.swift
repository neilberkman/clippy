import SwiftUI
import LaunchAtLogin

struct PreferencesView: View {
    @AppStorage("maxFilesShown") private var maxFilesShown: Int = 20
    @AppStorage("showFullPath") private var showFullPath: Bool = false
    
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
            }
            
            
            Section {
                HStack {
                    Button("Quit Draggy") {
                        NSApplication.shared.terminate(nil)
                    }
                    .foregroundColor(.red)
                    
                    Spacer()
                    
                    Text("v1.0.0")
                        .foregroundColor(.secondary)
                        .font(.caption)
                }
            }
        }
        .formStyle(.grouped)
        .frame(width: 400, height: 350)
    }
}