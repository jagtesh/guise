class Guise < Formula
  desc "Identity manager for developer tools (CLI TUI)"
  homepage "https://github.com/jagtesh/guise"
  url "https://github.com/jagtesh/guise/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "0000000000000000000000000000000000000000000000000000000000000000"
  license "BSD-3-Clause"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w")
  end

  test do
    assert_match "Guise", shell_output("#{bin}/guise --help", 1)
  end
end
