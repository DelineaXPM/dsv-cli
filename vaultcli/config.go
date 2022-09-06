package vaultcli

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	cst "thy/constants"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

/*
	Configuration file format:
	It is required that configuration file is written in "yaml" format. Even though,
	"Viper" supports different formats (e.g. "yaml", "json", etc.) out-of-the-box,
	the CLI uses "yaml" marshaling when creats a new file, appends to it or updates it.
*/

// Supported configuration file names.
const (
	legacyCliConfigName = ".thy.yml"
	cliConfigName       = ".dsv.yml"
)

// Supproted versions of configuration files.
const (
	v1 = "v1"
	v2 = "v2"
)

var (
	ErrFileNotFound = errors.New("configuration file not found")
)

type ConfigFile struct {
	DefaultProfile string

	version  string
	profiles map[string]map[string]interface{}

	path          string // Sets path to configuration file.
	isDefaultPath bool   // Denotes whether path was set by user or is a default one.
	raw           []byte // Raw content of the file.
}

// configFileFormatV1 defines first (initial) version of the CLI configuration file.
// Initially config file was just a list of profiles. By default profile named "default"
// is used.
type configFileFormatV1 map[string]map[string]interface{}

// configFileFormatV2 defines second version of the CLI configuration file.
// This version of the config file allows to change default profile referencing by name
// any existing profile from "profiles".
type configFileFormatV2 struct {
	Version        string                            `yaml:"version"`
	DefaultProfile string                            `yaml:"defaultProfile"`
	Profiles       map[string]map[string]interface{} `yaml:"profiles"`
}

func ViperInit(cfgFile string, profile string, args []string) error {
	viper.SetEnvPrefix(cst.EnvVarPrefix)
	envReplacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(envReplacer)
	viper.AutomaticEnv()

	cf, err := ReadConfigFile(cfgFile)
	if err != nil {
		return err
	}

	if profile == "" {
		profile = viper.GetString(cst.Profile)
		if profile == "" {
			profile = cf.DefaultProfile
		}
	}

	// Set profile name to lower case globally.
	profile = strings.ToLower(profile)
	viper.Set(cst.Profile, profile)

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

func LookupConfigPath(dir string) string {
	path := filepath.Join(dir, cliConfigName)
	_, err := os.Stat(path)
	if err == nil {
		return path
	}

	legacyPath := filepath.Join(dir, legacyCliConfigName)
	_, err = os.Stat(legacyPath)
	if err == nil {
		// Return legacy path only if exists.
		return legacyPath
	}

	return path
}

func NewConfigFile(path string) (*ConfigFile, error) {
	isDefaultPath := path == ""

	if isDefaultPath {
		home, err := homedir.Dir()
		if err != nil {
			return nil, fmt.Errorf("failed to determine home directory: %w", err)
		}
		path = LookupConfigPath(home)
	}

	return &ConfigFile{
		path:          path,
		isDefaultPath: isDefaultPath,
		profiles:      make(map[string]map[string]interface{}),
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
	return &Profile{Name: p, data: data}, true
}

func (cf *ConfigFile) SetProfile(p *Profile) {
	if cf.profiles == nil {
		cf.profiles = make(map[string]map[string]interface{})
	}
	cf.profiles[p.Name] = p.data

	if len(cf.profiles) == 1 {
		cf.DefaultProfile = p.Name
	}
}

func (cf *ConfigFile) ListProfiles() []*Profile {
	p := make([]*Profile, 0, len(cf.profiles))
	for name, val := range cf.profiles {
		p = append(p, &Profile{Name: name, data: val})
	}
	sort.Slice(p, func(i, j int) bool {
		return strings.ToLower(p[i].Name) <= strings.ToLower(p[j].Name)
	})
	return p
}

func (cf *ConfigFile) ListProfilesNames() []string {
	p := make([]string, 0, len(cf.profiles))
	for name := range cf.profiles {
		p = append(p, name)
	}
	sort.Strings(p)
	return p
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
	if err == ErrFileNotFound {
		return err
	}
	if err != nil {
		return fmt.Errorf("could not read configuration file: %v", err)
	}
	return nil
}

func (cf *ConfigFile) Save() error {
	if cf.version == v1 && cf.isDefaultPath {
		cf.path = strings.Replace(cf.path, legacyCliConfigName, cliConfigName, 1)
	}

	log.Printf("[config] Saving configuration file at path %s.", cf.path)

	err := cf.save()
	if err != nil {
		return fmt.Errorf("could not save configuration file: %v", err)
	}
	return nil
}

func (cf *ConfigFile) Delete() error {
	log.Printf("[config] Deleting configuration file at path %s.", cf.path)

	err := os.Remove(cf.path)
	if err != nil {
		return fmt.Errorf("could not delete configuration file: %v", err)
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

	var dataMap map[string]interface{}
	err = yaml.Unmarshal(bytes, &dataMap)
	if err != nil {
		return fmt.Errorf("failed to unmarshal CLI configuration file: %v", err)
	}

	if ver, ok := dataMap["version"]; !ok {
		cf.version = v1
	} else {
		cf.version, ok = ver.(string)
		if !ok {
			return fmt.Errorf("invalid version: unexpected type %T", ver)
		}
	}

	profiles := make(map[string]map[string]interface{})
	switch cf.version {
	case v1:
		var ok bool
		for k, v := range dataMap {
			profiles[k], ok = v.(map[string]interface{})
			if !ok {
				return fmt.Errorf("failed to read profiles: invalid profile %s", k)
			}
		}
		cf.DefaultProfile = cst.DefaultProfile

	case v2:
		rawProfiles, ok := dataMap["profiles"].(map[string]interface{})
		if !ok {
			return errors.New("invalid configration file: missing profiles or profiles defined in unexpected format")
		}
		for k, v := range rawProfiles {
			profiles[k], ok = v.(map[string]interface{})
			if !ok {
				return fmt.Errorf("failed to read profiles: invalid profile %s", k)
			}
		}
		prof, ok := dataMap["defaultProfile"]
		if !ok {
			return errors.New("invalid configration file: missing defaultProfile")
		}
		cf.DefaultProfile, ok = prof.(string)
		if !ok {
			return fmt.Errorf("invalid defaultProfile: unexpected type %T", prof)
		}

	default:
		return fmt.Errorf("unsupported version: %s", cf.version)
	}

	cf.raw = bytes
	cf.profiles = profiles
	return nil
}

func (cf *ConfigFile) save() error {
	fileData := configFileFormatV2{
		Version:        v2,
		DefaultProfile: cf.DefaultProfile,
		Profiles:       cf.profiles,
	}

	dataYml, err := yaml.Marshal(fileData)
	if err != nil {
		return err
	}
	if err := os.WriteFile(cf.path, dataYml, 0600); err != nil {
		return err
	}
	return nil
}
