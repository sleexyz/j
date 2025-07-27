{
  description = "j - Modern justfile runner for monorepos";
  
  # Set the flake name for profile installation
  nixConfig = {
    flake-registry = "https://github.com/NixOS/flake-registry/raw/master/flake-registry.json";
  };
  
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };
  
  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        j = pkgs.callPackage ./j-go.nix { };
      in {
        packages = {
          default = j;
          j = j;
        };
        
        # Development shell for working on j
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            just
            git
            installShellFiles
          ];
          
          shellHook = ''
            echo "j development environment loaded"
            echo "Available commands: just build, just test, just install, just reinstall"
          '';
        };
      }
    );
}