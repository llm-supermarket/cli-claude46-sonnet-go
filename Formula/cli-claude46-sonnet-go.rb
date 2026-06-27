class CliClaude46SonnetGo < Formula
  desc "CLI tool for rclone-compatible file encryption and decryption"
  homepage "https://github.com/llm-supermarket/cli-claude46-sonnet-go"
  version "0.0.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/llm-supermarket/cli-claude46-sonnet-go/releases/download/v0.0.0/cli-claude46-sonnet-go-darwin-arm64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    else
      url "https://github.com/llm-supermarket/cli-claude46-sonnet-go/releases/download/v0.0.0/cli-claude46-sonnet-go-darwin-amd64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/llm-supermarket/cli-claude46-sonnet-go/releases/download/v0.0.0/cli-claude46-sonnet-go-linux-arm64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    else
      url "https://github.com/llm-supermarket/cli-claude46-sonnet-go/releases/download/v0.0.0/cli-claude46-sonnet-go-linux-amd64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    end
  end

  def install
    bin.install "cli-claude46-sonnet-go-darwin-arm64" => "cli-claude46-sonnet-go" if OS.mac? && Hardware::CPU.arm?
    bin.install "cli-claude46-sonnet-go-darwin-amd64" => "cli-claude46-sonnet-go" if OS.mac? && !Hardware::CPU.arm?
    bin.install "cli-claude46-sonnet-go-linux-arm64" => "cli-claude46-sonnet-go" if OS.linux? && Hardware::CPU.arm?
    bin.install "cli-claude46-sonnet-go-linux-amd64" => "cli-claude46-sonnet-go" if OS.linux? && !Hardware::CPU.arm?
  end

  test do
    assert_match "cli-claude46-sonnet-go version #{version}", shell_output("#{bin}/cli-claude46-sonnet-go version")
  end
end
