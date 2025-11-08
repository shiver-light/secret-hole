import SwiftUI

@main
struct SecretHoleApp: App {
  @StateObject private var appVM = AppViewModel()
  @StateObject private var heartbeat = HeartbeatManager.shared

  var body: some Scene {
    WindowGroup {
      RootView()
        .environmentObject(appVM)
        .onAppear {
          heartbeat.bind(appVM: appVM)
          heartbeat.start()
        }
    }
  }
}

struct RootView: View {
  @EnvironmentObject var appVM: AppViewModel
  var body: some View {
    Group {
      if appVM.userId != nil {
        HomeView()
      } else {
        AuthView()
      }
    }
  }
}
