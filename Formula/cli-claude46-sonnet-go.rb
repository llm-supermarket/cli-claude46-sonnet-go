class CliClaude46SonnetGo < Formula
  desc "CLI tool for rclone-compatible file encryption and decryption"
  homepage "https://github.com/llm-supermarket/cli-claude46-sonnet-go"
  version "1.0.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/llm-supermarket/cli-claude46-sonnet-go/releases/download/v1.0.0/cli-claude46-sonnet-go-darwin-arm64.tar.gz"
      sha256 "e55dfd69bb4290c1a3ae2bf01c19b69649e18526a8eb4b61284776e308359526"
    else
      url "https://github.com/llm-supermarket/cli-claude46-sonnet-go/releases/download/v1.0.0/cli-claude46-sonnet-go-darwin-amd64.tar.gz"
      sha256 "e620dfbb7f7df40e8729ceb708bb6a6dfa91c47676a7b05cf332cddc8e5852a3"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/llm-supermarket/cli-claude46-sonnet-go/releases/download/v1.0.0/cli-claude46-sonnet-go-linux-arm64.tar.gz"
      sha256 "805027b430b8e9a2abf371f347ca02705891c0a2517ce0abc8356e8e17e2137d"
    else
      url "https://github.com/llm-supermarket/cli-claude46-sonnet-go/releases/download/v1.0.0/cli-claude46-sonnet-go-linux-amd64.tar.gz"
      sha256 "7f12fccb91fa5696dadd600814d5c5578a1f1f0ce363db99bc655ea652323b42"
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