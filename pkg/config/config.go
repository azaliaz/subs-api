package config

import (
	"log"
	"os"

	"github.com/caarlos0/env/v10"

	"gopkg.in/yaml.v3"
)

func ReadConfig(configFile string, cfg any) error {
	if configFile != "" && configFile != "none" {
		err := parseYaml(configFile, cfg)
		if err != nil {
			return err
		}
	} else {
		err := parseEnv(cfg)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseEnv(cfg any) error {
	err := env.Parse(cfg)
	if err != nil {
		return err
	}

	return nil
}

func parseYaml(file string, cfg any) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("failed to close file: %v", err)
		}
	}()

	yamlDecoder := yaml.NewDecoder(f)
	err = yamlDecoder.Decode(cfg)
	if err != nil {
		return err
	}

	return nil
}
