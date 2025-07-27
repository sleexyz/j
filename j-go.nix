# nix/j/j-go.nix
{ lib
, buildGoModule
, installShellFiles
, just
, git
}:

buildGoModule {
  pname = "j";
  version = "1.0.3";

  src = ./.;

  vendorHash = "sha256-m5mBubfbXXqXKsygF5j7cHEY+bXhAMcXUts5KBKoLzM=";

  subPackages = [ "cmd" ];

  postInstall = ''
    mv $out/bin/cmd $out/bin/j
    
    # Generate dynamic zsh completion script
    cat > _j_dynamic << 'EOF'
#compdef j

# Dynamic zsh completion for j - loads completion on demand
# This avoids materializing a static completion file

# Ensure compinit is loaded
autoload -U compinit
compinit

# Load dynamic completion from j binary
if command -v j >/dev/null 2>&1; then
  source <(j completion zsh)
else
  # Fallback if j is not in PATH
  return 1
fi
EOF
    
    # Install the dynamic completion script
    installShellCompletion --zsh _j_dynamic
  '';

  nativeBuildInputs = [ installShellFiles just git ];

  ldflags = [
    "-s"
    "-w"
    "-X main.version=1.0.3"
  ];

  # Skip tests that require network or git repository
  checkFlags = [
    "-skip=TestIntegration"
  ];

  meta = with lib; {
    description = "Modern justfile runner for monorepos";
    homepage = "https://github.com/websim/j";
    license = licenses.mit;
    maintainers = [ ];
    mainProgram = "j";
  };
}
