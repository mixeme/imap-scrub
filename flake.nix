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
        vendorHash = "sha256-P0p5mB0eT01u1AK3E6O5eQQUuCA5d9i069XNQQPi4Q4=";
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
