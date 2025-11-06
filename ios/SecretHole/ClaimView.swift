import SwiftUI

struct ClaimView: View {
  let userID: Int64
  @Environment(\.dismiss) var dismiss

  @State private var content = ""
  @State private var expiresAt = Date()
  @State private var now = Date()
  @State private var loading = true

  let timer = Timer.publish(every: 1, on: .main, in: .common).autoconnect()

  var body: some View {
    VStack(alignment: .leading, spacing: 16) {
      if loading {
        ProgressView("领取中…")
      } else {
        Text(content)
          .font(.title3)
          .privacySensitive()
          .textSelection(.disabled)
      }
      let remain = max(0, Int(expiresAt.timeIntervalSince(now)))
      ProgressView(value: Double(remain), total: Double(max(remain, 1)))
      Text("将在 \(remain) 秒后消失")
      Spacer()
    }
    .padding()
    .navigationTitle("听一句")
    .onAppear { Task { await claim() } }
    .onReceive(timer) { t in
      now = t
      if expiresAt <= t { dismiss() }
    }
  }

  func claim() async {
    loading = true
    defer { loading = false }
    do {
      let resp = try await API.claim(userID: userID)
      content = resp.body
      expiresAt = resp.expires_at
    } catch {
      content = "暂无可领取内容"
      expiresAt = Date().addingTimeInterval(5)
    }
  }
}
