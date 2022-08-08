package vaultcli

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	cst "thy/constants"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

/*
	Configuration file structure:
	The CLI configuration is a "profile". One configuration file can store many profiles.

	Configuration file format:
	It is required that configuration file is written in "yaml" format. Even though,
	"Viper" supports different formats (e.g. "yaml", "json", etc.) out-of-the-box,
	the CLI uses "yaml" marshaling when creats a new file, appends to it or updates it.
*/

var (
	ErrFileNotFound = errors.New("configuration file not found")
)

type ConfigFile struct {
	path       string
	raw        []byte
	profiles   map[string]map[string]interface{}
	newProfile map[string]map[string]interface{}
}

func ViperInit(cfgFile string, profile string, args []string) error {
	viper.SetEnvPrefix(cst.EnvVarPrefix)
	envReplacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(envReplacer)
	viper.AutomaticEnv()

	if profile == "" {
		profile = viper.GetString(cst.Profile)
		if profile == "" {
			profile = cst.DefaultProfile
		}
	}

	cf, err := ReadConfigFile(cfgFile)
	if err != nil {
		return err
	}

	config, ok := cf.GetProfile(profile)
	if !ok {
		return fmt.Errorf("profile %q not found in configuration file %q", profile, cf.GetPath())
	}

	err = viper.MergeConfigMap(config.data)
	if err != nil {
		return fmt.Errorf("cannot initialize Viper: %w", err)
	}

	return nil
}

func ReadConfigFile(path string) (*ConfigFile, error) {
	cf, err := NewConfigFile(path)
	if err != nil {
		return nil, err
	}
	err = cf.Read()
	if err != nil {
		return nil, err
	}
	return cf, nil
}

func DeleteConfigFile(path string) error {
	cf, err := NewConfigFile(path)
	if err != nil {
		return err
	}
	err = cf.Delete()
	if err != nil {
		return err
	}
	return nil
}

func NewConfigFile(path string) (*ConfigFile, error) {
	if path == "" {
		home, err := homedir.Dir()
		if err != nil {
			return nil, fmt.Errorf("failed to determine home directory: %w", err)
		}
		path = filepath.Join(home, cst.CliConfigName)
	}

	return &ConfigFile{
		path:     path,
		profiles: make(map[string]map[string]interface{}),
	}, nil
}

func (cf *ConfigFile) GetPath() string {
	return cf.path
}

func (cf *ConfigFile) Bytes() []byte {
	return cf.raw
}

func (cf *ConfigFile) GetProfile(p string) (*Profile, bool) {
	data, ok := cf.profiles[p]
	if !ok {
		return nil, false
	}
	return &Profile{name: p, data: data}, true
}

func (cf *ConfigFile) AddProfile(p *Profile) {
	raw := map[string]map[string]interface{}{p.name: p.data}

	if len(cf.profiles) != 0 {
		cf.newProfile = raw
	} else {
		cf.profiles = raw
	}
}

func (cf *ConfigFile) UpdateProfile(p *Profile) {
	cf.profiles[p.name] = p.data
}

func (cf *ConfigFile) RawUpdate(b []byte) error {
	var dataMap map[string]map[string]interface{}
	err := yaml.Unmarshal(b, &dataMap)
	if err != nil {
		return fmt.Errorf("invalid configuration: %v", err)
	}

	cf.raw = b
	cf.profiles = dataMap
	return nil
}

func (cf *ConfigFile) Read() error {
	log.Printf("[config] Reading configuration file at path %s.", cf.path)

	err := cf.read()
	if err != nil {
		return err
	}
	return nil
}

func (cf *ConfigFile) Save() error {
	log.Printf("[config] Saving configuration file at path %s.", cf.path)

	var err error
	if len(cf.newProfile) != 0 {
		err = cf.append()
	} else {
		err = cf.overwrite()
	}

	return err
}

func (cf *ConfigFile) Delete() error {
	log.Printf("[config] Deleting configuration file at path %s.", cf.path)

	err := os.Remove(cf.path)
	if err != nil {
		return err
	}

	return nil
}

func (cf *ConfigFile) read() error {
	_, err := os.Stat(cf.path)
	if err != nil {
		return ErrFileNotFound
	}

	bytes, err := os.ReadFile(cf.path)
	if err != nil {
		return err
	}

	var dataMap map[string]map[string]interface{}
	err = yaml.Unmarshal(bytes, &dataMap)
	if err != nil {
		return fmt.Errorf("failed to unmarshal CLI configuration file: %v", err)
	}

	cf.raw = bytes
	cf.profiles = dataMap
	return nil
}

func (cf *ConfigFile) append() error {
	dataYml, err := yaml.Marshal(cf.newProfile)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(cf.path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(dataYml); err != nil {
		return err
	}
	return nil
}

func (cf *ConfigFile) overwrite() error {
	dataYml, err := yaml.Marshal(cf.profiles)
	if err != nil {
		return err
	}
	if err := os.WriteFile(cf.path, dataYml, 0600); err != nil {
		return err
	}
	return nil
}
