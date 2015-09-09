class Gondor < Formula
  desc "Gondor CLI"
  homepage "https://next.gondor.io"
  head "https://github.com/eldarion-gondor/cli.git"
  url "https://github.com/eldarion-gondor/cli/archive/v0.3.2.tar.gz"
  sha256 "a77e29a3cf734c3f84225a908bfe8a83003976bf32352dd53f2a1ff00515f333"

  depends_on "go" => :build

  def install
    ENV["CGO_ENABLED"] = "0"
    system "make", "build"
    bin.install "bin/gondor"
  end

  test do
    system "#{bin}/gondor", "--version"
  end
end
