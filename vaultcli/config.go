package vaultcli

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	cst "github.com/DelineaXPM/dsv-cli/constants"

	"github.com/mitchellh/go-homedir"
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
//
//nolint:varnamelen // the length is good enough.
const (
	v1 = "v1"
	v2 = "v2"
)

var ErrFileNotFound = errors.New("configuration file not found")

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
	version, defaultProfile, profiles, err := parseRawConfig(b)
	if err != nil {
		return err
	}

	cf.version = version
	cf.DefaultProfile = defaultProfile
	cf.profiles = profiles
	cf.raw = b
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

	version, defaultProfile, profiles, err := parseRawConfig(bytes)
	if err != nil {
		return err
	}

	cf.version = version
	cf.DefaultProfile = defaultProfile
	cf.profiles = profiles
	cf.raw = bytes
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
	if err := os.WriteFile(cf.path, dataYml, 0o600); err != nil {
		return err
	}
	return nil
}

// parseRawConfig reads configuration bytes and returns version, default profile and list of profiles.
func parseRawConfig(b []byte) (string, string, map[string]map[string]interface{}, error) {
	if len(b) == 0 {
		return "", "", nil, nil
	}

	var dataMap map[string]interface{}
	err := yaml.Unmarshal(b, &dataMap)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to unmarshal CLI configuration: %v", err)
	}

	version := ""
	if ver, ok := dataMap["version"]; !ok {
		version = v1
	} else {
		version, ok = ver.(string)
		if !ok {
			return "", "", nil, fmt.Errorf("invalid version: unexpected type %T", ver)
		}
	}

	defaultProfile := ""
	profiles := make(map[string]map[string]interface{})
	switch version {
	case v1:
		var ok bool
		for k, v := range dataMap {
			profiles[k], ok = v.(map[string]interface{})
			if !ok {
				return "", "", nil, fmt.Errorf("failed to read profiles: invalid profile %s", k)
			}
		}
		defaultProfile = cst.DefaultProfile

	case v2:
		rawProfiles, ok := dataMap["profiles"].(map[string]interface{})
		if !ok {
			return "", "", nil, errors.New("invalid configration file: missing profiles or profiles defined in unexpected format")
		}
		for k, v := range rawProfiles {
			profiles[k], ok = v.(map[string]interface{})
			if !ok {
				return "", "", nil, fmt.Errorf("failed to read profiles: invalid profile %s", k)
			}
		}
		prof, ok := dataMap["defaultProfile"]
		if !ok {
			return "", "", nil, errors.New("invalid configration file: missing defaultProfile")
		}
		defaultProfile, ok = prof.(string)
		if !ok {
			return "", "", nil, fmt.Errorf("invalid defaultProfile: unexpected type %T", prof)
		}

	default:
		return "", "", nil, fmt.Errorf("unsupported version: %s", version)
	}

	if _, ok := profiles[defaultProfile]; !ok {
		return "", "", nil, fmt.Errorf("default profile %q is not defined", defaultProfile)
	}

	return version, defaultProfile, profiles, nil
}
