package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// HTTPServer groups settings related to the HTTP server (address, ports, TLS, etc.).
// We only have Addr for now, but grouping helps keep configuration organized.
type HTTPServer struct {
	// `yaml:"addr"` tells the YAML parser (cleanenv in our case) which YAML key
	// maps to this field. When the YAML file contains `http_server:\n  addr: ...`
	// it will fill this Addr value.
	Addr string `yaml:"addr"`
}

// Config is the root configuration structure for the application.
// Fields are annotated with tags that cleanenv understands for loading
// from YAML files and environment variables.
//
// Tags used here:
//   - yaml: name of the key when reading a YAML config file
//   - env: environment variable name to read if provided
//   - env-required: when set to "true" cleanenv will fail if the env var is missing
//   - env-default: a default used when neither YAML nor environment provide a value
//
// Example YAML that matches this struct:
// env: production
// storage_path: /var/data/app
// http_server:
//
//	addr: ":8080"
type Config struct {
	Env         string     `yaml:"env" env:"ENV" env-required:"true" env-default:"production"`
	StoragePath string     `yaml:"storage_path" env:"STORAGE_PATH" env-required:"true"`
	HTTPServer  HTTPServer `yaml:"http_server"`
}

// MustLoad loads configuration using the following precedence:
// 1) If CONFIG_PATH environment variable is set, that path is used.
// 2) Otherwise it looks for a -config flag passed to the program (CLI flag).
// 3) If neither is present, the function fatally exits the program.
//
// After a path is determined, the function checks the file exists and uses
// cleanenv to parse the YAML file into the Config struct. If any critical
// error occurs, the function logs the error and terminates the program.
//
// The function returns a pointer to a fully-populated Config on success.
func MustLoad() *Config {
	// Step A: try to read CONFIG_PATH environment variable first. This is useful
	// in containerized deployments or when an operator prefers environment-based
	// configuration.
	configPath := os.Getenv("CONFIG_PATH")

	// Step B: if CONFIG_PATH is empty, fallback to reading the command-line flag.
	//
	// Why use the flag package here?
	//  - flag.String returns a *string that will hold the flag value after flag.Parse()
	//  - This allows the program to accept `-config /path/to/config.yaml` at startup
	//  - Using flags is convenient for local development and for scripts
	if configPath == "" {
		// We store the pointer returned by flag.String in a variable named flagPtr.
		// Note: we do NOT call flag.Parse() until after we've declared all flags.
		flagPtr := flag.String("config", "", "path to the configuration file")
		// Parse parses the command-line flags from os.Args. It must be called
		// before we try to use the flag values (we dereference the pointer after
		// Parse). If you don't call Parse, flag values will remain at their defaults.
		flag.Parse()

		// Dereference the pointer to get the actual config path string.
		configPath = *flagPtr

		// If still empty, we cannot proceed because we don't know where to load the
		// configuration from. MustLoad is designed to fail-fast in this case.
		if configPath == "" {
			log.Fatal("config path is not set; set CONFIG_PATH or pass -config")
		}
	}

	// Step C: verify that the configuration file exists at the provided path.
	// os.Stat returns file info and an error. If the error indicates "file does
	// not exist" then we stop early with a helpful message.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	// Step D: read the configuration file into our Config struct.
	// cleanenv.ReadConfig supports YAML/JSON/TOML (depending on usage) and
	// additionally can populate values from environment variables defined by
	// struct tags. We delegate parsing and validation to cleanenv since it
	// provides convenient features (env-required, env-default, etc.).
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		// Any error reading/parsing the config is considered fatal. The error
		// message is included so operators can diagnose issues quickly.
		log.Fatalf("cannot read config file: %s", err.Error())
	}

	// Return a pointer to the populated configuration.
	return &cfg
}

/*
Additional notes and rationale (why things are done this way):

1) Why prefer environment variables (CONFIG_PATH) before CLI flags?
   - In containers and many deployment systems, environment variables are the
     standard way to inject configuration. Checking CONFIG_PATH first allows
     operators to override the CLI at runtime without changing startup scripts.

2) Why use flag.Parse() and a CLI flag at all?
   - CLI flags are convenient for developers running the program locally or
     in scripts. They make testing different configurations easy without
     changing the environment.

3) Why check file existence with os.Stat?
   - It gives a clear, early error message if the file is missing. Trying to
     read a non-existent file without checking might produce a less clear
     parsing error later.

4) Why use cleanenv and struct tags?
   - cleanenv simplifies reading configuration by supporting multiple sources
     (YAML + environment variables) and validating required values via tags.
     The struct tags document the expected keys and environment variables and
     make the wiring explicit.

5) Why does MustLoad exit the program on error (using log.Fatal/log.Fatalf)?
   - Configuration is critical: if required values (like STORAGE_PATH) are
     missing the program probably can't operate correctly. Failing fast and
     loudly helps avoid undefined behavior later on.

6) Suggested improvements (optional):
   - Add better validation for fields that need constraints (e.g. ensure
     StoragePath is writable, ensure HTTPServer.Addr is a valid address).
   - Support default file paths (e.g. look for ./config.yaml or /etc/myapp/config.yaml)
     before failing entirely.
   - Return an error instead of calling log.Fatal in library code, and let the
     caller choose whether to exit — this makes the package more reusable.
*/
