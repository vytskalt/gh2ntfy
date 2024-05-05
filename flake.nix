{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11";
    gomod2nix.url = "github:nix-community/gomod2nix";
  };

  outputs = { nixpkgs, gomod2nix, ... }: let
    forAllSystems = function:
      nixpkgs.lib.genAttrs [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ] (system:
        function (import nixpkgs {
          inherit system;
          overlays = [
            (import "${gomod2nix}/overlay.nix")
          ];
        }));
  in {
    packages = forAllSystems (pkgs: rec {
      package = pkgs.buildGoApplication {
        pname = "gh2ntfy";
        version = "0.1.0";
        pwd = ./.;
        src = ./.;
        modules = ./gomod2nix.toml;
      };

      image = pkgs.dockerTools.buildLayeredImage {
        name = "ghcr.io/vytskalt/gh2ntfy";
        tag = package.version;
        config.Cmd = [ "${package}/bin/gh2ntfy" ];
      };
    });

    devShells = forAllSystems (pkgs: {
      default = pkgs.mkShell {
        buildInputs = [
          pkgs.go
          pkgs.gomod2nix
        ];
      };
    });
  };
}
