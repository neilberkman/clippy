import SwiftUI
import Combine

// Interface-layer view model that bridges Core to SwiftUI
class ClipboardViewModel: ObservableObject {
    @Published var files: [ClipboardFile] = []
    @Published var isRefreshing = false
    
    // User preferences (interface concern)
    @AppStorage("maxFilesShown") var maxFilesShown: Int = 20
    
    private let monitor: ClipboardMonitor
    
    init(monitor: ClipboardMonitor = SystemClipboardMonitor()) {
        self.monitor = monitor
        
        // Bridge core events to UI updates
        monitor.onChange = { [weak self] files in
            DispatchQueue.main.async {
                self?.updateFiles(files)
            }
        }
        
        monitor.startMonitoring()
    }
    
    func refresh() {
        isRefreshing = true
        monitor.refresh()
        
        // UI feedback
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) { [weak self] in
            self?.isRefreshing = false
        }
    }
    
    private func updateFiles(_ newFiles: [ClipboardFile]) {
        // Apply interface-specific constraints
        let limitedFiles = Array(newFiles.prefix(maxFilesShown))
        
        files = limitedFiles
    }
    
}