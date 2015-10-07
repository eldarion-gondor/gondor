class Gondor < Formula
  desc "Gondor CLI"
  homepage "https://next.gondor.io"
  head "https://github.com/eldarion-gondor/cli.git"
  url "https://github.com/eldarion-gondor/cli/archive/v0.6.0.tar.gz"
  sha256 "1e8c297d1a17e6d826d85776f8215219bc9c8db1f45cb5bec220ce1d42a1beac"

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
