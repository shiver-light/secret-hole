import Foundation

struct AuthAnonReq: Codable { let device_hash: String }
struct AuthAnonResp: Codable {
  let user_id: Int64
  let date: String
  let sent_left: Int
  let recv_left: Int
}
struct PostReq: Codable {
  let body: String
  let read_duration: Int
}
struct PostResp: Codable { let message_id: Int64 }
struct ClaimResp: Codable {
  let message_id: Int64
  let body: String
  let read_duration: Int
  let read_at: Date
  let expires_at: Date
}
