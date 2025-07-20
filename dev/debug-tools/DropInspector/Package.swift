// swift-tools-version: 5.9

import PackageDescription

let package = Package(
    name: "DropInspector",
    platforms: [.macOS(.v13)],
    products: [
        .executable(name: "DropInspector", targets: ["DropInspector"])
    ],
    targets: [
        .executableTarget(
            name: "DropInspector",
            path: ".",
            sources: ["DropInspector.swift"]
        )
    ]
)
