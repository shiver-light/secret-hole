import SwiftUI

struct ComposeView: View {
  let userID: Int64
  @Environment(\.dismiss) var dismiss
  @State private var bodyText = ""
  @State private var dur: Double = 30
  @State private var sending = false

  var body: some View {
    Form {
      Section("内容") {
        TextEditor(text: $bodyText)
          .frame(minHeight: 160)
          .privacySensitive()
          .onChange(of: bodyText) { newValue in
            if newValue.count > 500 {
              bodyText = String(newValue.prefix(500))
            }
          }
      }

      Section("阅后时长：\(Int(dur)) 秒") {
        Slider(value: $dur, in: 1...120, step: 1)
      }

      Button {
        Task {
          sending = true
          defer { sending = false }
          do {
            _ = try await API.postMessage(
              userID: userID,
              body: bodyText.trimmingCharacters(in: .whitespacesAndNewlines),
              dur: Int(dur)
            )
            dismiss()
          } catch {
            print("post error", error)
          }
        }
      } label: {
        Text(sending ? "发送中…" : "发送")
      }
      .disabled(bodyText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
    }
    .navigationTitle("说一句")
  }
}
