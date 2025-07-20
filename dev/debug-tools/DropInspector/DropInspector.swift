import SwiftUI
import UniformTypeIdentifiers

@main
struct SafeDropInspectorApp: App {
    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}

struct ContentView: View {
    @State private var dropInfo: String = "Drop something here to inspect it..."
    @State private var isTargeted = false

    var body: some View {
        VStack {
            Text("Safe Drop Inspector")
                .font(.largeTitle)
                .padding()

            ScrollView {
                Text(dropInfo)
                    .font(.system(.body, design: .monospaced))
                    .textSelection(.enabled)
                    .padding()
                    .frame(maxWidth: .infinity, alignment: .leading)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .background(
                RoundedRectangle(cornerRadius: 10)
                    .fill(isTargeted ? Color.blue.opacity(0.2) : Color.gray.opacity(0.1))
            )
            .onDrop(of: [.item], isTargeted: $isTargeted) { providers in
                dropInfo = "=== Drop received ===\n\n"

                // Just list provider types WITHOUT loading data
                dropInfo += "PROVIDER TYPES (in order):\n"
                for (i, provider) in providers.enumerated() {
                    dropInfo += "\nProvider \(i+1):\n"
                    for (j, type) in provider.registeredTypeIdentifiers.enumerated() {
                        dropInfo += "  \(j+1). \(type)\n"
                    }
                }

                // Check pasteboard without loading data
                dropInfo += "\n\nPASTEBOARD TYPES (in order):\n"
                let pb = NSPasteboard(name: .drag)
                if let types = pb.types {
                    for (i, type) in types.enumerated() {
                        dropInfo += "  \(i+1). \(type.rawValue)\n"
                    }
                }

                return true
            }
        }
        .padding()
        .frame(width: 600, height: 500)
    }
}
