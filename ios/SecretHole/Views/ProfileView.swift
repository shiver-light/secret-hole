import SwiftUI

struct ProfileView: View {
  @EnvironmentObject var appVM: AppViewModel

  var body: some View {
    NavigationStack {
      Form {
        Section("账户") {
          Text("User ID: \(appVM.userId ?? 0)")
        }
        Section("说明") {
          Text("昵称（≤7 字）、头像（Base64）、颜色（#RRGGBB）后续版本支持。")
            .font(.footnote)
        }
      }
      .navigationTitle("我")
    }
  }
}
