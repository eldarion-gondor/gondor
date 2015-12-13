class Gondor < Formula
  desc "Gondor CLI"
  homepage "https://next.gondor.io"
  head "https://github.com/eldarion-gondor/cli.git"
  url "https://github.com/eldarion-gondor/cli/archive/v0.7.0.tar.gz"
  sha256 "04e8eaf31e0e5fb7b8846ea4f3947647fc41b7b8e43a6d13c25ce8c4a7483e6c"

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
