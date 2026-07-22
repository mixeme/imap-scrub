class ImapScrub < Formula
  desc "Reduce IMAP mailbox size via configurable search-and-action rules"
  homepage "https://github.com/mixeme/imap-scrub"
  url "https://github.com/mixeme/imap-scrub/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "1547ca1ec53cb6f2c61bac886bfce15bea8dad0e60af2753be2b7b6a57b141b1"
  license "MIT"
  head "https://github.com/mixeme/imap-scrub.git", branch: "develop"

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X main.appVersion=#{version}"
    system "go", "build", *std_go_args(ldflags: ldflags)
  end

  test do
    output = shell_output("#{bin}/imap-scrub --version")
    assert_match "imap-scrub", output
  end
end
