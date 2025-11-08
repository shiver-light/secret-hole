import Foundation

struct AnonAuthRequest: Codable {
  let device_hash: String
}
struct AnonAuthResponse: Codable {
  let user_id: Int64
}

struct QuotaResponse: Codable {
  let sent: Int
  let recv: Int
}

struct PostMessageRequest: Codable {
  let body: String
  let read_duration: Int  // 秒
}

struct ClaimResponse: Codable {
  let id: Int64
  let author_id: Int64
  let body: String
  let read_duration: Int
  let expires_at: String?  // ISO8601，可选
}

struct AckDeleteResponse: Codable { let ok: Bool }

struct HeartbeatRequest: Codable { let is_foreground: Bool }
struct HeartbeatResponse: Codable { let ok: Bool }

struct ReplyRequest: Codable { let body: String }
struct ReplyResponse: Codable {
  let ok: Bool
  let pending: Bool?
  let active_until: String?
  let streak: Int?
}

struct DMSendRequest: Codable {
  let to: Int64
  let body: String
}
struct DMSendResponse: Codable {
  let ok: Bool
  let active_until: String?
}

enum APIError: Error, LocalizedError {
  case invalidURL
  case badStatus(Int)
  case decoding
  case server(String)
  case notLoggedIn
  var errorDescription: String? {
    switch self {
    case .invalidURL: return "Invalid URL"
    case .badStatus(let c): return "HTTP \(c)"
    case .decoding: return "Decoding error"
    case .server(let s): return "Server: \(s)"
    case .notLoggedIn: return "Not logged in"
    }
  }
}

extension ISO8601DateFormatter {
  static let withFraction: ISO8601DateFormatter = {
    let f = ISO8601DateFormatter()
    f.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
    return f
  }()
}
