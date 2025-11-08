import Combine

import Foundation

@MainActor
final class AppViewModel: ObservableObject {
  @Published var userId: Int64?
  @Published var quota: QuotaResponse?
  @Published var lastClaim: ClaimResponse?
  @Published var tempFriendActiveUntil: Date?
  @Published var tempFriendPeerId: Int64?
  @Published var tempFriendStreak: Int = 0

  // 用于 DM 倒计时
  var remainingSeconds: Int {
    guard let until = tempFriendActiveUntil else { return 0 }
    return max(Int(until.timeIntervalSinceNow), 0)
  }

  func loginAnon(deviceHash: String) async throws {
    let res = try await APIClient.shared.authAnon(deviceHash: deviceHash)
    self.userId = res.user_id
    try? await refreshQuota()
  }

  func refreshQuota() async throws {
    guard let uid = userId else { return }
    quota = try await APIClient.shared.getQuota(userId: uid)
  }

  func postPublic(text: String, duration: Int) async throws {
    guard let uid = userId else { return }
    _ = try await APIClient.shared.postPublicMessage(userId: uid, text: text, duration: duration)
    try? await refreshQuota()
  }

  func claimOne() async throws -> ClaimResponse {
    guard let uid = userId else { throw APIError.notLoggedIn }
    let res = try await APIClient.shared.claimMessage(userId: uid)
    self.lastClaim = res
    return res
  }

  func ackDelete(messageId: Int64) async {
    guard let uid = userId else { return }
    _ = try? await APIClient.shared.ackDelete(userId: uid, messageId: messageId)
  }

  func replyTo(messageId: Int64, text: String) async throws -> ReplyResponse {
    guard let uid = userId else { throw APIError.notLoggedIn }
    let res = try await APIClient.shared.reply(userId: uid, messageId: messageId, text: text)
    if let untilStr = res.active_until,
      let untilDate = ISO8601DateFormatter.withFraction.date(from: untilStr)
        ?? ISO8601DateFormatter().date(from: untilStr)
    {
      tempFriendActiveUntil = untilDate
      // 推断 peer id：从 lastClaim 或 UI 传入更稳，这里简化
      if let claim = lastClaim { tempFriendPeerId = claim.author_id }
      tempFriendStreak = res.streak ?? tempFriendStreak
    }
    return res
  }

  func sendDM(to peer: Int64, text: String) async throws {
    guard let uid = userId else { throw APIError.notLoggedIn }
    let res = try await APIClient.shared.sendDM(userId: uid, to: peer, text: text)
    if let untilStr = res.active_until {
      if let untilDate = ISO8601DateFormatter.withFraction.date(from: untilStr)
        ?? ISO8601DateFormatter().date(from: untilStr)
      {
        tempFriendActiveUntil = untilDate
        tempFriendPeerId = peer
      }
    }
  }

  func clearTempFriendIfExpired() {
    if remainingSeconds <= 0 {
      tempFriendActiveUntil = nil
      tempFriendPeerId = nil
    }
  }
}
