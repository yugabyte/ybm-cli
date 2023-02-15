# typed: false
# frozen_string_literal: true

require_relative "lib/custom_download_strategy"
class YbmCli < Formula
  desc "YugabyteDB Managed CLI"
  homepage "https://github.com/yugabyte/ybm-cli"
  version "0.0.5"
  license "Apache"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/yugabyte/ybm-cli/releases/download/v0.0.5/ybm_0.0.5_darwin_x86_64.zip", using: GitHubPrivateRepositoryReleaseDownloadStrategy
      sha256 "871890402221c8cdde1f8ba1c87f0d15981dce1029983ccc07b347f2e1425f72"

      def install
        bin.install "ybm-cli"
      end
    end
    if Hardware::CPU.arm?
      url "https://github.com/yugabyte/ybm-cli/releases/download/v0.0.5/ybm_0.0.5_darwin_arm64.zip", using: GitHubPrivateRepositoryReleaseDownloadStrategy
      sha256 "aa4c2fb3bf4362be7615b225ebc475c2947dbe27a2b90c4d041cb205735ce204"

      def install
        bin.install "ybm-cli"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/yugabyte/ybm-cli/releases/download/v0.0.5/ybm_0.0.5_linux_x86_64.zip", using: GitHubPrivateRepositoryReleaseDownloadStrategy
      sha256 "290cf8a1929a115c34818d2c7bf02048261dc75b0fd9420a031f3f415f669499"

      def install
        bin.install "ybm-cli"
      end
    end
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/yugabyte/ybm-cli/releases/download/v0.0.5/ybm_0.0.5_linux_arm64.zip", using: GitHubPrivateRepositoryReleaseDownloadStrategy
      sha256 "46f844efed7a7acf931a2e78e8d2f8335f8c3b5a291e1eb52a0ed1464f339f2d"

      def install
        bin.install "ybm-cli"
      end
    end
  end
end
