{
  description = "github.com/brnsampson/ezconf development and build flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs = { self, nixpkgs }:
  let 
    systems = [
      "aarch64-linux"
      "x86_64-linux"
      "aarch64-darwin"
      "x86_64-darwin"
    ];
    inherit (nixpkgs) lib;
    forEachSystem = lib.genAttrs systems;
    pkgsFor = system: nixpkgs.legacyPackages.${system};
  in
  {

    devShells = forEachSystem (
      system:
      let
        pkgs = pkgsFor system;
      in
      {
        default = import ./shell.nix { inherit pkgs; };
      }
    );

    packages = forEachSystem (
      system:
      let
        pkgs = pkgsFor system;
        local = import ./default.nix { pkgs }
      in
      {
        default = local.ezconf
      }
    );
  };
}
