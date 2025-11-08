import Foundation

final class APIClient {
  static let shared = APIClient()
  private init() {}

  private func req(_ path: String, method: String = "GET", body: Data? = nil, userId: Int64?)
    -> URLRequest
  {
    var url = Env.baseURL
    url.append(path: path)
    var r = URLRequest(url: url)
    r.httpMethod = method
    r.setValue(Env.json, forHTTPHeaderField: "content-type")
    if let uid = userId {
      r.setValue(String(uid), forHTTPHeaderField: "X-User-ID")
    }
    r.httpBody = body
    return r
  }

  private func run<T: Decodable>(_ request: URLRequest, _ type: T.Type) async throws -> T {
    let (data, resp) = try await URLSession.shared.data(for: request)
    guard let http = resp as? HTTPURLResponse else { throw APIError.invalidURL }
    guard 200..<300 ~= http.statusCode else {
      // 尝试解析 {"error":"..."}
      if let obj = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
        let msg = obj["error"] as? String
      {
        throw APIError.server(msg)
      }
      throw APIError.badStatus(http.statusCode)
    }
    do {
      return try JSONDecoder().decode(T.self, from: data)
    } catch {
      throw APIError.decoding
    }
  }

  // MARK: - API

  func authAnon(deviceHash: String) async throws -> AnonAuthResponse {
    let body = try JSONEncoder().encode(AnonAuthRequest(device_hash: deviceHash))
    let req = req("/v1/auth/anon", method: "POST", body: body, userId: nil)
    return try await run(req, AnonAuthResponse.self)
  }

  func getQuota(userId: Int64) async throws -> QuotaResponse {
    let req = req("/v1/me/quota", userId: userId)
    return try await run(req, QuotaResponse.self)
  }

  func postPublicMessage(userId: Int64, text: String, duration: Int) async throws
    -> AckDeleteResponse
  {
    let body = try JSONEncoder().encode(PostMessageRequest(body: text, read_duration: duration))
    let req = req("/v1/messages", method: "POST", body: body, userId: userId)
    return try await run(req, AckDeleteResponse.self)
  }

  func claimMessage(userId: Int64) async throws -> ClaimResponse {
    let req = req("/v1/messages/claim", method: "POST", body: Data(), userId: userId)
    return try await run(req, ClaimResponse.self)
  }

  func ackDelete(userId: Int64, messageId: Int64) async throws -> AckDeleteResponse {
    let req = req(
      "/v1/messages/\(messageId)/ack-delete", method: "POST", body: Data(), userId: userId)
    return try await run(req, AckDeleteResponse.self)
  }

  func heartbeat(userId: Int64, isForeground: Bool) async throws {
    let body = try JSONEncoder().encode(HeartbeatRequest(is_foreground: isForeground))
    let req = req("/v1/presence/heartbeat", method: "POST", body: body, userId: userId)
    _ = try await run(req, HeartbeatResponse.self)
  }

  func reply(userId: Int64, messageId: Int64, text: String) async throws -> ReplyResponse {
    let body = try JSONEncoder().encode(ReplyRequest(body: text))
    let req = req("/v1/messages/\(messageId)/reply", method: "POST", body: body, userId: userId)
    return try await run(req, ReplyResponse.self)
  }

  func sendDM(userId: Int64, to peer: Int64, text: String) async throws -> DMSendResponse {
    let body = try JSONEncoder().encode(DMSendRequest(to: peer, body: text))
    let req = req("/v1/dm/send", method: "POST", body: body, userId: userId)
    return try await run(req, DMSendResponse.self)
  }
}
