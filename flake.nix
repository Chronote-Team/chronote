{
  description = "Chronote-golang";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          go          # go 编译器
          gopls       # go 语言服务器
          gotools     # go imports 等工具
          jq          # 命令行 json 处理
          git
          curl
        ];

        shellHook = ''
          export GOPATH=$PWD/.gopath
          export PATH=$GOPATH/bin:$PATH
        '';
      };
    };
}
