class Gondor < Formula
  desc "Gondor CLI"
  homepage "https://next.gondor.io"
  head "https://github.com/eldarion-gondor/cli.git"
  url "https://github.com/eldarion-gondor/cli/archive/v0.5.2.tar.gz"
  sha256 "bfef093a30115ac8d627ebbd3695cb15c6d742e7c9771633dfa9fb3f6da55a07"

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
