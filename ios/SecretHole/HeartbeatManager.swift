import Combine

import Foundation

import UIKit

final class HeartbeatManager: ObservableObject {
  static let shared = HeartbeatManager()
  private init() {
    NotificationCenter.default.addObserver(
      self, selector: #selector(appActive), name: UIApplication.didBecomeActiveNotification,
      object: nil)
    NotificationCenter.default.addObserver(
      self, selector: #selector(appInactive), name: UIApplication.willResignActiveNotification,
      object: nil)
  }

  private weak var appVM: AppViewModel?
  private var timer: Timer?
  private var isForeground = false

  func bind(appVM: AppViewModel) { self.appVM = appVM }

  @objc private func appActive() {
    isForeground = true
    tick()
    startTimer()
  }

  @objc private func appInactive() {
    isForeground = false
    tick()
    stopTimer()
  }

  func start() {
    isForeground = UIApplication.shared.applicationState == .active
    tick()
    startTimer()
  }

  private func startTimer() {
    timer?.invalidate()
    timer = Timer.scheduledTimer(withTimeInterval: 10, repeats: true) { [weak self] _ in
      self?.tick()
    }
  }
  private func stopTimer() {
    timer?.invalidate()
    timer = nil
  }

  private func tick() {
    guard let uid = appVM?.userId else { return }
    Task { try? await APIClient.shared.heartbeat(userId: uid, isForeground: isForeground) }
  }
}
