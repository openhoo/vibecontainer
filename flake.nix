{
  description = "vibecontainer Go CLI development flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachSystem
      [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ]
      (
        system:
        let
          pkgs = import nixpkgs { inherit system; };
          pname = "vibecontainer";
          version = "0.0.0";
          commitRev = if self ? shortRev then self.shortRev else "dirty";
          vibecontainer = pkgs.buildGoModule {
            inherit pname version;
            src = ./.;
            subPackages = [ "cmd/vibecontainer" ];
            vendorHash = "sha256-sjjjB24LQoCqw7l/s9gdOQ1RrYIt7ELSoevTKQszhTA=";
            ldflags = [
              "-s"
              "-w"
              "-X main.version=${version}"
              "-X main.commit=${commitRev}"
              "-X main.date=1970-01-01T00:00:00Z"
            ];
            doCheck = false;
          };
        in
        {
          packages = {
            inherit vibecontainer;
            default = vibecontainer;
          };

          apps = {
            vibecontainer = {
              type = "app";
              program = "${vibecontainer}/bin/vibecontainer";
            };
            default = self.apps.${system}.vibecontainer;
          };

          devShells.default = pkgs.mkShell {
            packages = with pkgs; [
              go
              gopls
              golangci-lint
              gotools
              delve
              docker-client
              docker-compose
              git
              nixfmt
            ];
          };

          formatter = pkgs.nixfmt;

          checks = {
            vibecontainer-build = vibecontainer;
          };
        }
      );
}
