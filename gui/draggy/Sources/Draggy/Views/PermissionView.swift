import SwiftUI
import AppKit

struct PermissionView: View {
    let folders: [String]
    @Binding var isPresented: Bool
    
    var body: some View {
        VStack(spacing: 16) {
            Image(systemName: "folder.badge.questionmark")
                .font(.system(size: 48))
                .foregroundColor(.orange)
            
            Text("Permission Required")
                .font(.headline)
            
            Text("Draggy needs permission to access your folders to show recent downloads.")
                .font(.subheadline)
                .foregroundColor(.secondary)
                .multilineTextAlignment(.center)
                .padding(.horizontal)
            
            VStack(alignment: .leading, spacing: 8) {
                Text("Folders that need access:")
                    .font(.caption)
                    .foregroundColor(.secondary)
                
                ForEach(folders, id: \.self) { folder in
                    HStack {
                        Image(systemName: "folder")
                            .foregroundColor(.blue)
                        Text(folder)
                            .font(.caption)
                            .lineLimit(1)
                            .truncationMode(.middle)
                    }
                }
            }
            .padding()
            .background(Color.gray.opacity(0.1))
            .cornerRadius(8)
            
            HStack(spacing: 12) {
                Button("Cancel") {
                    isPresented = false
                }
                .buttonStyle(.plain)
                
                Button("Grant Access") {
                    openSystemPreferences()
                }
                .buttonStyle(.borderedProminent)
            }
            
            Text("You may need to restart Draggy after granting permission.")
                .font(.caption2)
                .foregroundColor(.secondary)
        }
        .padding()
        .frame(width: 350)
    }
    
    private func openSystemPreferences() {
        if let url = URL(string: "x-apple.systempreferences:com.apple.preference.security?Privacy_AllFiles") {
            NSWorkspace.shared.open(url)
        }
    }
}