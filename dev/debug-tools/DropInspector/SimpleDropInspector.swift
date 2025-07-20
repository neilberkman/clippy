import SwiftUI
import UniformTypeIdentifiers

@main
struct SimpleDropInspectorApp: App {
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
            Text("Simple Drop Inspector")
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

                for (index, provider) in providers.enumerated() {
                    dropInfo += "Item \(index + 1):\n"
                    dropInfo += "Registered types:\n"

                    // Just list the types, don't load the data yet
                    for typeID in provider.registeredTypeIdentifiers {
                        dropInfo += "  - \(typeID)\n"
                    }

                    dropInfo += "\nFirst 3 types in detail:\n"

                    // Only check first 3 types to avoid overload
                    for typeID in provider.registeredTypeIdentifiers.prefix(3) {
                        dropInfo += "\n[\(typeID)]:\n"

                        // Check if it's an image type
                        if typeID.contains("image") || typeID == "public.tiff" || typeID == "public.png" {
                            provider.loadDataRepresentation(forTypeIdentifier: typeID) { data, error in
                                DispatchQueue.main.async {
                                    if let data = data {
                                        dropInfo += "  ✓ Image data: \(data.count) bytes\n"
                                    } else {
                                        dropInfo += "  ✗ No image data\n"
                                    }
                                }
                            }
                        }

                        // Check if it's a file URL
                        if typeID.contains("file") || typeID == "public.file-url" {
                            provider.loadDataRepresentation(forTypeIdentifier: typeID) { data, error in
                                DispatchQueue.main.async {
                                    if let data = data, let str = String(data: data, encoding: .utf8) {
                                        dropInfo += "  File URL: \(str)\n"
                                    }
                                }
                            }
                        }
                    }
                }

                return true
            }
        }
        .padding()
        .frame(width: 600, height: 500)
    }
}
