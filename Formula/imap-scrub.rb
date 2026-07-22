class ImapScrub < Formula
  desc "Reduce IMAP mailbox size via configurable search-and-action rules"
  homepage "https://github.com/mixeme/imap-scrub"
  license "MIT"
  version "0.1.0"

  # Stable installs use the prebuilt binaries already published on the GitHub
  # release (see .github/workflows/build-release.yml) — no Go toolchain needed.
  on_macos do
    on_arm do
      url "https://github.com/mixeme/imap-scrub/releases/download/v0.1.0/imap-scrub-darwin-arm64.tar.gz"
      sha256 "34ec9011351d93f7e9cff1cf4100b975c6d81a10951426ec1d5b58fec8deb65a"
    end
    on_intel do
      url "https://github.com/mixeme/imap-scrub/releases/download/v0.1.0/imap-scrub-darwin-amd64.tar.gz"
      sha256 "f271aeff705cae651f01f031eca7cc0ac828e6884a3bb122c93fcc4402db39c8"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/mixeme/imap-scrub/releases/download/v0.1.0/imap-scrub-linux-amd64.tar.gz"
      sha256 "96b16710ab6406c0accb2191de6e3308272ecd77383a282b22439c70e7283f0a"
    end
    on_arm do
      url "https://github.com/mixeme/imap-scrub/releases/download/v0.1.0/imap-scrub-linux-arm64.tar.gz"
      sha256 "245dc38ccabcd54af678fba705add4f2f8054de36882566f3b3b59b605abdc40"
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
