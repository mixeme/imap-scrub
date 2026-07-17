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
      version = "0.1.0";
      src = ./.;
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in
    {
      packages.${system}.default = pkgs.buildGoModule {
        inherit src pname version;
        # Recalculate after go.mod changes: nix build 2>&1 | grep got:
        vendorHash = "sha256-WkPG3HLCSEpJ6Sqz3+mF4X++D2MDKK34YIpc9oehhvI=";
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
