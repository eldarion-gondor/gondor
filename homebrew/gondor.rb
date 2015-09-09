class Gondor < Formula
  desc "Gondor CLI"
  homepage "https://next.gondor.io"
  head "https://github.com/eldarion-gondor/cli.git"
  url "https://github.com/eldarion-gondor/cli/archive/v0.2.3.tar.gz"
  sha256 "71923e6afc23583e1a8123f2272bd88e4d32e0c6043b77c8dcf27655e353e07d"

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
