class Gondor < Formula
  desc "Gondor CLI"
  homepage "https://next.gondor.io"
  head "https://github.com/eldarion-gondor/cli.git"
  url "https://github.com/eldarion-gondor/cli/archive/v0.5.0.tar.gz"
  sha256 "344dce25f087c60d2a428218f94854692dd99bea4268d3fb0603d308bdde7dc9"

  depends_on "go" => :build

  def install
    ENV["CGO_ENABLED"] = "0"
    system "make", "build"
    bin.install "bin/g3a"
  end

  test do
    system "#{bin}/g3a", "--version"
  end
end
