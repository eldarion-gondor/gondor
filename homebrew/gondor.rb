class Gondor < Formula
  desc "Gondor CLI"
  homepage "https://next.gondor.io"
  head "https://github.com/eldarion-gondor/cli.git"
  url "https://github.com/eldarion-gondor/cli/archive/v0.3.4.tar.gz"
  sha256 "737e2012e45bc96828cdf44cc5b6fc603ceaaf5919a4d15e773d000cbe21a231"

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
