{ pkgs ? import <nixpkgs> {}, ... }:
let
  lib = pkgs.lib;
in
{
  ezconf = pkgs.callPackage ./exconf.nix
}
