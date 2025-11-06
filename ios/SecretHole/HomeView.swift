import SwiftUI

struct HomeView: View {
  @Binding var userID: Int64
  @State private var sentLeft = 0
  @State private var recvLeft = 0

  var body: some View {
    VStack(spacing: 24) {
      Text("树洞：阅后即焚").font(.largeTitle).bold()
      HStack {
        VStack {
          Text("可说")
          Text("\(sentLeft)").font(.title)
        }
        Divider().frame(height: 40)
        VStack {
          Text("可听")
          Text("\(recvLeft)").font(.title)
        }
      }.padding().background(.thinMaterial).clipShape(RoundedRectangle(cornerRadius: 16))

      HStack {
        NavigationLink("说一句") { ComposeView(userID: userID) }
          .buttonStyle(.borderedProminent)
        NavigationLink("听一句") { ClaimView(userID: userID) }
          .buttonStyle(.bordered)
      }
      Spacer()
    }
    .padding()
    .task { await refresh() }
    .onAppear { Task { await refresh() } }
  }

  func refresh() async {
    guard userID != 0 else { return }
    do {
      let (s, r) = try await API.meQuota(userID: userID)
      sentLeft = s
      recvLeft = r
    } catch { print("quota error", error) }
  }
}
