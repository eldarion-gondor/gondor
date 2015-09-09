class Gondor < Formula
  desc "Gondor CLI"
  homepage "https://next.gondor.io"
  head "https://github.com/eldarion-gondor/cli.git"
  url "https://github.com/eldarion-gondor/cli/archive/v0.2.8.tar.gz"
  sha256 "5cdd153b0ab2177f54ed33caecb2f7bdfc0856c7824fa1550f67aafcc318086a"

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
