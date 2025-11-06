import Foundation

enum API {
  static var BASE_URL = URL(string: "http://127.0.0.1:8080")!

  static func authAnon(deviceHash: String) async throws -> AuthAnonResp {
    var req = URLRequest(url: BASE_URL.appending(path: "/v1/auth/anon"))
    req.httpMethod = "POST"
    req.addValue("application/json", forHTTPHeaderField: "Content-Type")
    req.httpBody = try JSONEncoder().encode(AuthAnonReq(device_hash: deviceHash))
    let (data, _) = try await URLSession.shared.data(for: req)
    return try JSONDecoder().decode(AuthAnonResp.self, from: data)
  }

  static func meQuota(userID: Int64) async throws -> (Int, Int) {
    var req = URLRequest(url: BASE_URL.appending(path: "/v1/me/quota"))
    req.addValue(String(userID), forHTTPHeaderField: "X-User-ID")
    let (data, _) = try await URLSession.shared.data(for: req)
    let obj = try JSONSerialization.jsonObject(with: data) as! [String: Any]
    return (obj["sent_left"] as! Int, obj["recv_left"] as! Int)
  }

  static func postMessage(userID: Int64, body: String, dur: Int) async throws -> Int64 {
    var req = URLRequest(url: BASE_URL.appending(path: "/v1/messages"))
    req.httpMethod = "POST"
    req.addValue("application/json", forHTTPHeaderField: "Content-Type")
    req.addValue(String(userID), forHTTPHeaderField: "X-User-ID")
    req.httpBody = try JSONEncoder().encode(PostReq(body: body, read_duration: dur))
    let (data, _) = try await URLSession.shared.data(for: req)
    return try JSONDecoder().decode(PostResp.self, from: data).message_id
  }

  static func claim(userID: Int64) async throws -> ClaimResp {
    var req = URLRequest(url: BASE_URL.appending(path: "/v1/messages/claim"))
    req.httpMethod = "POST"
    req.addValue(String(userID), forHTTPHeaderField: "X-User-ID")
    let (data, _) = try await URLSession.shared.data(for: req)
    let dec = JSONDecoder()
    dec.dateDecodingStrategy = .iso8601
    return try dec.decode(ClaimResp.self, from: data)
  }
}
