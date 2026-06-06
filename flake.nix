{
  description = "oarsman - a postman-like OpenAPI curl command generator for the command line";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        oarsman = pkgs.buildGoModule {
          pname = "oarsman";
          version = "0.1.0";

          src = ./.;

          vendorHash = "sha256-UnMAB8pDaoKhPuYvz9jLtwu5ssZ4LFJIpO4YQSgZ6Kk=";

          meta = with pkgs.lib; {
            description = "A postman-like OpenAPI curl command generator for the command line";
            homepage = "https://github.com/aistriketeam/oarsman";
            license = licenses.mit;
            mainProgram = "oarsman";
          };
        };
      in
      {
        packages.default = oarsman;
        packages.oarsman = oarsman;

        apps.default = {
          type = "app";
          program = "${oarsman}/bin/oarsman";
        };

        devShells.default = pkgs.mkShell {
          packages = [ pkgs.go pkgs.gopls pkgs.gotools ];
        };
      });
}
