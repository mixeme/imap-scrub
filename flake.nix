{
  description = "build imap-scrub";

  inputs = {
    nixpkgs.url = "github:kompismoln/nixpkgs/nixos-unstable";
  };

  outputs =
    {
      self,
      nixpkgs,
    }:
    let
      pname = "imap-scrub";
      version = "0.3.0";
      src = ./.;
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in
    {
      packages.${system}.default = pkgs.buildGoModule {
        inherit src pname version;
        # Stale since golang.org/x/oauth2 was added — run `nix build` once and
        # paste the reported `got:` hash here.
        vendorHash = nixpkgs.lib.fakeHash;
      };

      devShells.${system} = {
        default = pkgs.mkShell {
          name = "${pname}-dev";
          packages = with pkgs; [
            go
          ];
        };
      };
    };
}
