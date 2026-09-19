class ImapScrub < Formula
  desc "Reduce IMAP mailbox size via configurable search-and-action rules"
  homepage "https://github.com/mixeme/imap-scrub"
  license "MIT"
  version "0.3.0"

  # Stable installs use the prebuilt binaries already published on the GitHub
  # release (see .github/workflows/build-release.yml) — no Go toolchain needed.
  on_macos do
    on_arm do
      url "https://github.com/mixeme/imap-scrub/releases/download/v0.3.0/imap-scrub-darwin-arm64.tar.gz"
      sha256 "da70ddbbdc28d8ffba30ccefeb8ed7b08a1312f7cfe1f99a216fafd248797f4e"
    end
    on_intel do
      url "https://github.com/mixeme/imap-scrub/releases/download/v0.3.0/imap-scrub-darwin-amd64.tar.gz"
      sha256 "6fe6738aaafb9a616d08baee226b63f53f3fdebd5217ef2a6e16ac81a717af73"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/mixeme/imap-scrub/releases/download/v0.3.0/imap-scrub-linux-amd64.tar.gz"
      sha256 "896eacdb3d2742d9d5f32d2cc163d1656e27963fabc1a8b5206a0aaa32a302e0"
    end
    on_arm do
      url "https://github.com/mixeme/imap-scrub/releases/download/v0.3.0/imap-scrub-linux-arm64.tar.gz"
      sha256 "419559658d22628ca057529cdcdc5a14f70cdfca27b33a2e5bb7beb6ebd425f4"
    end
  end

  # --HEAD builds the unreleased develop branch from source instead.
  head "https://github.com/mixeme/imap-scrub.git", branch: "develop"
  depends_on "go" => :build if build.head?

  def install
    if build.head?
      ldflags = "-s -w -X main.appVersion=#{version}"
      system "go", "build", *std_go_args(ldflags: ldflags)
    else
      bin.install "imap-scrub"
    end
  end

  test do
    output = shell_output("#{bin}/imap-scrub --version")
    assert_match "imap-scrub", output
  end
end
