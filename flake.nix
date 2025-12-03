{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };
  outputs = {
    self,
    nixpkgs,
  }: let
    system = "x86_64-linux"; # or "aarch64-darwin" for M1/M2 Macs
    pkgs = import nixpkgs {inherit system;};
  in {
    devShells.${system}.default = pkgs.mkShell {
      buildInputs = with pkgs; [
        go
        golangci-lint
        wails
        gnumake
        nodejs_24
        zig
        clang
      ];
      shellHook = ''
        go env -w GOPATH=$HOME/.local/share/go
        export NPM_CONFIG_PREFIX="$HOME/.local"
        export CGO_ENABLED=1
        export PATH="$HOME/.local/bin:$PATH"
        export PATH="$HOME/.local/share/go/bin:$PATH"
      '';
    };
  };
}
