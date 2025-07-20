import Foundation

// Core domain logic for time formatting
struct TimeFormatting {

    static func formatRelativeTime(from date: Date) -> String {
        let now = Date()
        let interval = now.timeIntervalSince(date)

        if interval < 60 {
            return "\(Int(interval))s ago"
        } else if interval < 3600 {
            return "\(Int(interval / 60))m ago"
        } else if interval < 86400 {
            return "\(Int(interval / 3600))h ago"
        } else {
            return "\(Int(interval / 86400))d ago"
        }
    }

    static func formatAbsoluteTime(from date: Date) -> String {
        let formatter = DateFormatter()
        formatter.dateFormat = "MMM d HH:mm"
        return formatter.string(from: date)
    }
}
