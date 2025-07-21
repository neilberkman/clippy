import SwiftUI

struct UpdateNotificationView: View {
    @ObservedObject var updateChecker: UpdateChecker

    var body: some View {
        HStack {
            Image(systemName: "arrow.down.circle.fill")
                .foregroundColor(.accentColor)

            VStack(alignment: .leading, spacing: 2) {
                Text("Update Available")
                    .font(.caption)
                    .fontWeight(.semibold)

                Text(updateChecker.updateMessage)
                    .font(.caption2)
                    .foregroundColor(.secondary)
            }

            Spacer()

            HStack(spacing: 8) {
                Button("Later") {
                    updateChecker.dismissUpdate()
                }
                .buttonStyle(.plain)
                .font(.caption)
                .foregroundColor(.secondary)

                if updateChecker.isInstalledViaHomebrew() {
                    Button("Copy") {
                        // Copy brew command to clipboard
                        let command = updateChecker.getBrewUpdateCommand()
                        NSPasteboard.general.clearContents()
                        NSPasteboard.general.setString(command, forType: .string)
                        updateChecker.dismissUpdate()
                    }
                    .buttonStyle(.borderedProminent)
                    .font(.caption)
                    .controlSize(.small)
                    .help("Copy update command to clipboard")
                } else {
                    Button("Download") {
                        updateChecker.openDownloadPage()
                        updateChecker.dismissUpdate()
                    }
                    .buttonStyle(.borderedProminent)
                    .font(.caption)
                    .controlSize(.small)
                }
            }
        }
        .padding(.horizontal)
        .padding(.vertical, 8)
        .background(Color(NSColor.windowBackgroundColor).opacity(0.9))
    }
}
