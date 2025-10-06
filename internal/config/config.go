// Copyright © 2025 Steve Taranto staranto@gmail.com
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"gopkg.in/yaml.v3"
)

type Type struct {
	Source    string
	Namespace string
	Data      map[string]interface{}
}

var Config Type

func init() {
	_, _ = Load()
}

func Load(cfgFilePath ...string) (Type, error) {
	path, err := getConfigPath()
	if err != nil {
		return Type{}, err
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return Type{}, err
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(bytes, &data); err != nil {
		return Type{}, err
	}

	Config = Type{
		Source: path,
		Data:   data}

	return Config, nil
}

// get traverses the map using a dotted key path
func (cfg *Type) get(kspec string) (any, error) {
	if len(cfg.Data) == 0 {
		_, _ = Load(cfg.Source)
	}

	candidateKeys := []string{"", kspec}
	if cfg.Namespace != "" {
		candidateKeys[0] = cfg.Namespace + "." + kspec
	}

	for _, key := range candidateKeys {
		keys := strings.Split(key, ".")
		var current interface{} = Config.Data

		success := true
		for _, key := range keys {
			m, ok := current.(map[string]interface{})
			if !ok {
				success = false
				break
			}
			current, ok = m[key]
			if !ok {
				success = false
				break
			}
		}

		if success {
			return current, nil
			// if str, ok := current.(string); ok {
			// 	return str, nil
			// }
			// return "", fmt.Errorf("value at path '%s' is not a string", key)
		}
	}

	return nil, fmt.Errorf("no valid path found among: %v", candidateKeys)
}

func GetString(key string, defaultValue ...string) (string, error) {
	if len(Config.Data) == 0 {
		_, _ = Load()
	}

	val, err := Config.get(key)
	if err != nil {
		if len(defaultValue) == 1 {
			return defaultValue[0], nil
		}
		return "", err
	}

	s, ok := val.(string)
	if !ok {
		return "", errors.New("value is not a string")
	}

	return s, nil
}

func GetInt(key string, defaultValue ...int) (int, error) {
	if len(Config.Data) == 0 {
		_, _ = Load()
	}

	val, err := Config.get(key)
	if err != nil && Config.Namespace != "" {
		val, err = Config.get(Config.Namespace + "." + key)
	}

	if err != nil {
		if len(defaultValue) == 1 {
			return defaultValue[0], nil
		}
		return 0, err
	}

	// YAML numbers may be unmarshaled as int/float64 depending on content.
	switch v := val.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, errors.New("value is not an int")
	}
}

func getConfigPath() (string, error) {

	var candidates []string = []string{
		os.Getenv("XDG_CONFIG_HOME"),
		os.Getenv("APPDATA"),
		os.Getenv("HOME"),
	}

	for _, c := range candidates {
		file := filepath.Join(c, "tfctl.yaml")
		if fileInfo, err := os.Stat(file); err == nil {
			if !fileInfo.IsDir() {
				log.Debugf("using config file: %s", file)
				return file, nil
			}
		}
	}
	return "", fmt.Errorf("no config file found in standard locations")
}
