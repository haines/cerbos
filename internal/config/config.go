// Copyright 2021-2022 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"go.uber.org/config"
)

var ErrConfigNotLoaded = errors.New("config not loaded")

var conf = &configHolder{}

type Section interface {
	Key() string
}

type Defaulter interface {
	SetDefaults()
}

type Validator interface {
	Validate() error
}

// Load loads the config file at the given path.
func Load(confFile string, overrides map[string]interface{}) error {
	finfo, err := os.Stat(confFile)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", confFile, err)
	}

	if finfo.IsDir() {
		return fmt.Errorf("config file path is a directory: %s", confFile)
	}

	return doLoad(config.File(confFile), config.Static(overrides))
}

func LoadReader(reader io.Reader, overrides map[string]interface{}) error {
	return doLoad(config.Source(reader), config.Static(overrides))
}

func LoadMap(m map[string]interface{}) error {
	return doLoad(config.Static(m))
}

func doLoad(sources ...config.YAMLOption) error {
	opts := append(sources, config.Expand(os.LookupEnv)) //nolint:gocritic
	provider, err := config.NewYAML(opts...)
	if err != nil {
		if strings.Contains(err.Error(), "couldn't expand environment") {
			return fmt.Errorf("error loading configuration due to unknown environment variable. Config values containing '$' are interpreted as environment variables. Use '$$' to escape literal '$' values: [%w]", err)
		}
		return fmt.Errorf("failed to load config: %w", err)
	}

	conf.replaceProvider(provider)

	return nil
}

// Get populates out with the configuration at the given key.
// Populate out with default values before calling this function to ensure sane defaults if there are any.
func Get(key string, out interface{}) error {
	return conf.Get(key, out)
}

// GetSection populates a config section.
func GetSection(section Section) error {
	return conf.Get(section.Key(), section)
}

type configHolder struct {
	provider config.Provider
	mu       sync.RWMutex
}

func (ch *configHolder) Get(key string, out interface{}) error {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if ch.provider == nil {
		if d, ok := out.(Defaulter); ok {
			d.SetDefaults()
			return nil
		}

		return ErrConfigNotLoaded
	}

	// set defaults if any are specified
	if d, ok := out.(Defaulter); ok {
		d.SetDefaults()
	}

	if err := ch.provider.Get(key).Populate(out); err != nil {
		return err
	}

	// validate if a validate function is available
	if v, ok := out.(Validator); ok {
		return v.Validate()
	}

	return nil
}

func (ch *configHolder) replaceProvider(provider config.Provider) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	ch.provider = provider
}
