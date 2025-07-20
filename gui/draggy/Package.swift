// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "Draggy",
    platforms: [
        .macOS(.v13)
    ],
    products: [
        .executable(
            name: "Draggy",
            targets: ["Draggy"]
        )
    ],
    dependencies: [
        .package(url: "https://github.com/sindresorhus/LaunchAtLogin", from: "5.0.0")
    ],
    targets: [
        .executableTarget(
            name: "Draggy",
            dependencies: ["LaunchAtLogin"],
            path: "Sources/Draggy",
            linkerSettings: [
                .unsafeFlags([
                    "-L.",  // Look for libraries in current directory
                    "-lclippy"  // Link with libclippy.a
                ])
            ]
        )
    ]
)
