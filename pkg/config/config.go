package config

import (
	"os"
	"path/filepath"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog/log"
)

type (
	Opts struct {
		Config    any
		Paths     []string
		Filenames []string
	}
)

func Load(opts Opts) error {
	// Track if any config file was loaded
	configLoaded := false
	var loadedFiles []string

	// Try to load specified config files
	for _, f := range opts.Filenames {
		for _, p := range opts.Paths {
			fp := filepath.Join(p, f)
			// Check if file exists
			if _, fileErr := os.Stat(fp); fileErr != nil {
				// File doesn't exist, skip it (don't return error)
				log.Debug().Str("file", fp).Msg("Config file not found, skipping")
				continue
			}
			// Try to read the config file
			if err := cleanenv.ReadConfig(fp, opts.Config); err != nil {
				// If file exists but can't be read, log warning and continue
				// This allows the app to fallback to env vars
				log.Warn().Str("file", fp).Err(err).Msg("Failed to parse config file, will use env vars")
				continue
			}
			configLoaded = true
			loadedFiles = append(loadedFiles, fp)
			log.Info().Str("file", fp).Msg("Config file loaded successfully")
		}
	}

	// Finally, read from environment variables (this will override file values if set)
	// This allows the app to work in containers with env vars only
	if err := cleanenv.ReadEnv(opts.Config); err != nil {
		return err
	}

	// Log configuration source
	if configLoaded {
		log.Info().Strs("files", loadedFiles).Msg("Configuration loaded from files")
		log.Info().Msg("Environment variables will override file values if set")
	} else {
		log.Info().Msg("No config files found, using environment variables only")
	}

	return nil
}
