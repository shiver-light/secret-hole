import SwiftUI

struct HomeView: View {
  var body: some View {
    TabView {
      PublicSendView()
        .tabItem { Label("发送", systemImage: "square.and.pencil") }

      PublicClaimView()
        .tabItem { Label("领取", systemImage: "tray") }

      DMChatView()
        .tabItem { Label("临时好友", systemImage: "bubble.left.and.bubble.right.fill") }

      ProfileView()
        .tabItem { Label("我", systemImage: "person.crop.circle") }
    }
  }
}
