import SwiftUI

struct PublicSendView: View {
  @EnvironmentObject var appVM: AppViewModel
  @State private var text: String = ""
  @State private var duration: Int = 30
  @State private var error: String?

  var body: some View {
    NavigationStack {
      Form {
        Section("今日配额（发送）") {
          HStack {
            Text("已用")
            Spacer()
            Text("\(appVM.quota?.sent ?? 0) / 3")
          }
          Button("刷新配额") { Task { try? await appVM.refreshQuota() } }
        }

        Section("发送公共消息") {
          TextField("说点什么…", text: $text, axis: .vertical)
          Stepper("阅读后持续 \(duration) 秒", value: $duration, in: 5...300, step: 5)

          Button("发送") {
            Task {
              do {
                try await appVM.postPublic(text: text, duration: duration)
                text = ""
                try? await appVM.refreshQuota()
              } catch {
                self.error = error.localizedDescription
              }
            }
          }
          .disabled(text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
        }

        if let err = error {
          Section { Text(err).foregroundStyle(.red) }
        }
      }
      .navigationTitle("发送")
    }
  }
}
