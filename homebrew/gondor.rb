class Gondor < Formula
  desc "Gondor CLI"
  homepage "https://next.gondor.io"
  head "https://github.com/eldarion-gondor/cli.git"
  url "https://github.com/eldarion-gondor/cli/archive/v0.6.1.tar.gz"
  sha256 "22434e33e8892354e82c76059e102fa47fa81d4d351eb28c0cbc591b0fae9740"

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
