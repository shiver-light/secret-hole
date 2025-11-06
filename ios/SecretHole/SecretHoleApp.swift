import SwiftUI

@main
struct SecretHoleApp: App {
  @State private var userID: Int64 = 0

  var body: some Scene {
    WindowGroup {
      NavigationStack {
        HomeView(userID: $userID)
      }
      .task {
        if userID == 0 {
          do {
            let resp = try await API.authAnon(
              deviceHash: UIDevice.current.identifierForVendor?.uuidString ?? "dev")
            userID = resp.user_id
          } catch {
            print("auth error", error)
          }
        }
      }
    }
  }
}
