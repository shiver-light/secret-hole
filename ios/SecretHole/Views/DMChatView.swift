import SwiftUI

struct DMChatView: View {
  @EnvironmentObject var appVM: AppViewModel
  @State private var input: String = ""
  @State private var tick: Timer?

  var body: some View {
    VStack(spacing: 8) {
      if let peer = appVM.tempFriendPeerId,
        let until = appVM.tempFriendActiveUntil
      {
        HStack {
          Text("与 \(peer) 临时好友中")
          Spacer()
          Text(remainingText(until: until))
            .monospacedDigit()
            .foregroundStyle(.green)
        }
        .padding(.horizontal)

        Spacer()

        HStack {
          TextField("发消息…", text: $input)
            .textFieldStyle(.roundedBorder)
          Button("发送") {
            Task {
              do {
                try await appVM.sendDM(to: peer, text: input)
                input = ""
              } catch { /* 可弹提示 */  }
            }
          }
          .disabled(
            appVM.remainingSeconds <= 0
              || input.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
        }
        .padding(.horizontal)
        .padding(.bottom)
      } else {
        ContentUnavailableView(
          "暂无临时好友", systemImage: "bubble.left.and.bubble.right",
          description: Text("在“领取”页对公共消息做一次回复，并在双方前台活跃时建立临时好友")
        )
        .padding(.top, 40)
      }
    }
    .navigationTitle("临时好友")
    .onAppear { startTick() }
    .onDisappear { stopTick() }
  }

  private func remainingText(until: Date) -> String {
    let left = max(Int(until.timeIntervalSinceNow), 0)
    let m = left / 60
    let s = left % 60
    return String(format: "%02d:%02d", m, s)
  }

  private func startTick() {
    stopTick()
    tick = Timer.scheduledTimer(withTimeInterval: 1, repeats: true) { _ in
      appVM.clearTempFriendIfExpired()
    }
  }
  private func stopTick() {
    tick?.invalidate()
    tick = nil
  }
}
