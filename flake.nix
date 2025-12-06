{
  description = "Tuberly - Web Server in Go";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };

        # Project configuration
        projectName = "tuberly";
        dbName = "${projectName}.db";
        migrationsDir = "./db/migrations";
        queriesDir = "./db/queries";

        # Setup script to create project structure and config files
        setup-project = pkgs.writeShellScriptBin "setup-project" ''
          echo "Setting up project structure..."

          # Create directories
          mkdir -p ${migrationsDir}
          mkdir -p ${queriesDir}

          # Create .air.toml if it doesn't exist
          if [ ! -f .air.toml ]; then
            cat > .air.toml << 'EOF'
          root = "."
          testdata_dir = "testdata"
          tmp_dir = "tmp"

          [build]
            args_bin = []
            bin = "./tmp/main"
            cmd = "go build -o ./tmp/main ."
            delay = 1000
            exclude_dir = ["assets", "tmp", "vendor", "testdata"]
            exclude_file = []
            exclude_regex = ["_test.go"]
            exclude_unchanged = false
            follow_symlink = false
            full_bin = ""
            include_dir = []
            include_ext = ["go", "tpl", "tmpl", "html"]
            include_file = []
            kill_delay = "0s"
            log = "build-errors.log"
            poll = false
            poll_interval = 0
            post_cmd = []
            pre_cmd = []
            rerun = false
            rerun_delay = 500
            send_interrupt = false
            stop_on_error = false

          [color]
            app = ""
            build = "yellow"
            main = "magenta"
            runner = "green"
            watcher = "cyan"

          [log]
            main_only = false
            time = false

          [misc]
            clean_on_exit = false

          [screen]
            clear_on_rebuild = false
            keep_scroll = true
          EOF
            echo "✓ Created .air.toml"
          fi

          # Create sqlc.yaml if it doesn't exist
          if [ ! -f sqlc.yaml ]; then
            cat > sqlc.yaml << 'EOF'
          version: "2"
          sql:
            - schema: "${migrationsDir}"
              queries: "${queriesDir}"
              engine: "sqlite"
              gen:
                go:
                  package: "database"
                  out: "internal/database"
                  emit_json_tags: true
                  emit_interface: false
                  emit_exact_table_names: false
          EOF
            echo "✓ Created sqlc.yaml"
          fi

          echo "✓ Project structure ready"
          echo "  Migrations: ${migrationsDir}"
          echo "  Queries: ${queriesDir}"
          echo "  Database: ${dbName}"
        '';

      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            # Go toolchain
            go
            golangci-lint
            gotools
            delve

            # Database
            sqlite
            goose # Go database migration tool
            sqlc # SQL -> Go code generation tool

            # Development tools
            air # Live reload for Go
            jq # JSON processing for API testing

            # HTTP testing
            httpie # Perform HTTP requests in the CLI

            # Media processing
            ffmpeg # Includes both ffmpeg and ffprobe

            # Cloud tools
            awscli2

            # Security
            govulncheck # Vulnerability scanner for Go dependencies

            # Helper scripts
            setup-project
          ];

          shellHook = ''
            # SQLite configuration
            export DB_PATH="$PWD/${dbName}"
            export DATABASE_URL="$DB_PATH"

            # Goose environment variables
            export GOOSE_DRIVER="sqlite3"
            export GOOSE_DBSTRING="$DB_PATH"
            export GOOSE_MIGRATION_DIR="${migrationsDir}"

            # Development environment variables
            export PLATFORM="dev" # dev | prod | test

            # Auto-setup project structure on first run
            if [ ! -d "${migrationsDir}" ] || [ ! -f "sqlc.yaml" ] || [ ! -f ".air.toml" ]; then
              setup-project
            fi

            echo "✓ ${projectName} dev environment ready"
            echo "  Database: ${dbName}"
            echo "  FFMPEG and AWS CLI available"
          '';
        };
      }
    );
}
