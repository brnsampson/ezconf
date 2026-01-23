{ pkgs ? import <nixpkgs> {}, ... }:
let
  lib = pkgs.lib;
in
{
  ezconf = pkgs.callPackage ./ezconf.nix
}
