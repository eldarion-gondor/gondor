class Gondor < Formula
  desc "Gondor CLI"
  homepage "https://next.gondor.io"
  head "https://github.com/eldarion-gondor/cli.git"
  url "https://github.com/eldarion-gondor/cli/archive/v0.3.5.tar.gz"
  sha256 "771abbdff59d12e6d24d8a1456a227bd2eaa31f629a67ae6a42c5a785e7cc0cf"

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
