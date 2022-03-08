// Copyright 2021-2022 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

//go:build !js
// +build !js

package config

import (
	"context"
	"fmt"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

// LoadAndWatch automatically reloads configuration if the config file changes.
func LoadAndWatch(ctx context.Context, confFile string, overrides map[string]interface{}) error {
	if err := Load(confFile, overrides); err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to start file watcher: %w", err)
	}

	if err := watcher.Add(confFile); err != nil {
		return fmt.Errorf("failed to add watch to %s: %w", confFile, err)
	}

	go func() {
		defer watcher.Close()

		log := zap.S().Named("config.watch").With("file", confFile)
		log.Info("Watching config file for changes")

		for {
			select {
			case <-ctx.Done():
				log.Info("Stopping config file watch")

				return
			case event, ok := <-watcher.Events:
				if !ok {
					log.Info("Stopping config file watch")
					return
				}

				switch {
				case event.Op&fsnotify.Create == fsnotify.Create:
					fallthrough
				case event.Op&fsnotify.Write == fsnotify.Write:
					if err := Load(confFile, overrides); err != nil {
						log.Warnw("Failed to reload config file", "error", err)
					} else {
						log.Info("Config file reloaded")
					}
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					log.Warn("Config file removed")
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					log.Warn("Config file renamed")
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					log.Info("Stopping config file watch")
					return
				}

				log.Warnw("Error watching config file", "error", err)
			}
		}
	}()

	return nil
}
