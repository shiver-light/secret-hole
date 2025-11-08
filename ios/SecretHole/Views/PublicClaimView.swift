import SwiftUI

struct PublicClaimView: View {
  @EnvironmentObject var appVM: AppViewModel

  @State private var claimed: ClaimResponse?
  @State private var replyText: String = ""
  @State private var error: String?

  @State private var countdown: Int = 0
  @State private var timer: Timer?

  var body: some View {
    NavigationStack {
      VStack(spacing: 12) {
        Form {
          Section("今日配额（领取）") {
            HStack {
              Text("已用")
              Spacer()
              Text("\(appVM.quota?.recv ?? 0) / 3")
            }
            HStack {
              Button("刷新配额") { Task { try? await appVM.refreshQuota() } }
              Spacer()
              Button {
                Task { await claimOne() }
              } label: {
                Label("领取一条消息", systemImage: "tray.and.arrow.down.fill")
              }
              .buttonStyle(.borderedProminent)
            }
          }

          Section("消息内容") {
            if let c = claimed {
              messageCard(c)
            } else {
              ContentUnavailableView("暂无消息", systemImage: "tray", description: Text("点击上方领取一条"))
            }
          }

          if let err = error {
            Section { Text(err).foregroundStyle(.red) }
          }
        }
      }
      .navigationTitle("领取")
    }
    .onDisappear { stopTimer() }
  }

  @ViewBuilder
  private func messageCard(_ c: ClaimResponse) -> some View {
    VStack(alignment: .leading, spacing: 12) {
      HStack {
        Text("来自用户 \(c.author_id)")
          .font(.subheadline)
          .foregroundStyle(.secondary)
        Spacer()
        Text(timeLeftText())
          .font(.system(.subheadline, design: .monospaced))
          .foregroundStyle(countdown > 10 ? .green : .orange)
      }

      Text(c.body)
        .font(.title3)
        .padding(.vertical, 4)

      Divider()

      // 一次回复
      VStack(alignment: .leading, spacing: 8) {
        Text("一次回复").font(.subheadline).foregroundStyle(.secondary)
        HStack {
          TextField("写点什么…（仅一次）", text: $replyText)
            .textFieldStyle(.roundedBorder)
          Button("发送") { Task { await replyOnce(c) } }
            .buttonStyle(.borderedProminent)
            .disabled(replyText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
        }
      }

      // 手动确认删除（可选，本地倒计时到 0 时也会自动删）
      HStack {
        Spacer()
        Button("已读并删除") { Task { await ackAndClear(c) } }
          .buttonStyle(.bordered)
      }
    }
    .padding(.vertical, 4)
    .onAppear { startCountdown(for: c) }
  }

  // MARK: - Actions

  private func claimOne() async {
    error = nil
    stopTimer()
    do {
      guard appVM.userId != nil else { throw APIError.notLoggedIn }
      let res = try await appVM.claimOne()
      claimed = res
      replyText = ""
      startCountdown(for: res)
    } catch {
      self.error = error.localizedDescription
    }
  }

  private func replyOnce(_ c: ClaimResponse) async {
    error = nil
    do {
      let res = try await appVM.replyTo(messageId: c.id, text: replyText)
      if res.ok {
        await appVM.ackDelete(messageId: c.id)
        clearLocal()
      }
    } catch {
      self.error = error.localizedDescription
    }
  }

  private func ackAndClear(_ c: ClaimResponse) async {
    await appVM.ackDelete(messageId: c.id)
    clearLocal()
  }

  private func clearLocal() {
    stopTimer()
    claimed = nil
    replyText = ""
    countdown = 0
  }

  // MARK: - Countdown

  private func startCountdown(for c: ClaimResponse) {
    stopTimer()
    if let expStr = c.expires_at,
      let exp = ISO8601DateFormatter.withFraction.date(from: expStr)
        ?? ISO8601DateFormatter().date(from: expStr)
    {
      countdown = max(Int(exp.timeIntervalSinceNow), 0)
    } else {
      countdown = max(c.read_duration, 0)
    }
    guard countdown > 0 else {
      Task { await ackAndClear(c) }
      return
    }
    timer = Timer.scheduledTimer(withTimeInterval: 1, repeats: true) { [weak self] _ in
      guard let self else { return }
      countdown = max(countdown - 1, 0)
      if countdown == 0 {
        Task { await ackAndClear(c) }
      }
    }
  }

  private func stopTimer() {
    timer?.invalidate()
    timer = nil
  }

  private func timeLeftText() -> String {
    let m = countdown / 60
    let s = countdown % 60
    return String(format: "%02d:%02d", m, s)
  }
}
