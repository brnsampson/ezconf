{ pkgs, lib, buildGoModule }:
buildGoModule( finalAttrs:
  {
  pname = "ezconf";
  version = "0.1.0";

  # openssh is just for the git fetch.
  nativeBuildInputs = [ openssh ];

  GOPRIVATE = "github.com/Dsperse/*";
  GIT_ASKPASS = "";
  src = builtins.fetchGit {
    url = "git@github.com:brnsampson/ezconf.git";
    ref = "refs/tags/v${finalAttrs.version}";
    rev = "";
  };

  vendorHash = lib.fakeHash;

  meta = {
    description = "A go generate tool for producing useful config loaders based on simple config structs.";
    homepage = "https://github.com/brnsampson/ezconf";
    license = lib.licenses.mit;
  };
})
