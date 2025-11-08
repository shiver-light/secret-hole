import SwiftUI

struct AuthView: View {
  @EnvironmentObject var appVM: AppViewModel
  @State private var deviceHash: String = "dev-\(UUID().uuidString.prefix(6))"
  @State private var error: String?

  var body: some View {
    VStack(spacing: 16) {
      Text("树洞 · 匿名登录").font(.largeTitle.bold())
      TextField("设备标识（可自定义）", text: $deviceHash)
        .textFieldStyle(.roundedBorder)
        .padding(.horizontal)

      Button("开始使用") {
        Task {
          do { try await appVM.loginAnon(deviceHash: deviceHash) } catch {
            self.error = error.localizedDescription
          }
        }
      }
      .buttonStyle(.borderedProminent)

      if let err = error {
        Text(err).foregroundStyle(.red)
      }
    }
    .padding()
  }
}
