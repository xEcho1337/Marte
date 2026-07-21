package config

import (
	"bufio"
	_ "embed"
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed default.yml
var defaultConfig []byte

type Config struct {
	ApiPort        int            `yaml:"api-port"`
	SubmitterPort  int            `yaml:"submitter-port"`
	Password       string         `yaml:"password"`
	TeamToken      string         `yaml:"team-token"`
	TeamId         int            `yaml:"your-team-id"`
	NOPTeamId      int            `yaml:"nop-team-id"`
	RoundTime      int            `yaml:"round-time"`
	FlagIdFormat   string         `yaml:"flag-id-format"`
	TeamCount      int            `yaml:"team-count"`
	FlagRegex      string         `yaml:"flag-regex"`
	TeamIpFormat   string         `yaml:"team-ip-format"`
	TeamIpFile     string         `yaml:"team-ip-file"`
	TeamIPs        []string       `yaml:"-"`
	UrlFlagIds     string         `yaml:"url-flag-ids"`
	UrlFlagSubmit  string         `yaml:"url-flag-submit"`
	FlagBatch      int            `yaml:"flag-batch"`
	FlagSubmitRate int            `yaml:"flag-submit-rate"`
	Submitter      string         `yaml:"submitter"`
	Services       map[string]int `yaml:"services"`
}

func Ensure(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	return os.WriteFile(path, defaultConfig, 0644)
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var config Config

	if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if err := ValidateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func ValidateConfig(cfg *Config) error {
	if cfg.Submitter != "cyber_challenge" && cfg.Submitter != "forcad" && cfg.Submitter != "enowars" {
		return errors.New("Invalid submitter: " + cfg.Submitter)
	}

	if cfg.FlagIdFormat != "cyber_challenge" && cfg.FlagIdFormat != "forcad" && cfg.FlagIdFormat != "enowars" {
		return errors.New("Invalid flag-id-format: " + cfg.FlagIdFormat)
	}

	if !strings.Contains(cfg.TeamIpFormat, "{}") && cfg.TeamIpFile == "" {
		return errors.New("Invalid team ip format, expecting {} inside: " + cfg.TeamIpFormat + " (or set team-ip-file)")
	}

	if cfg.TeamIpFile != "" {
		return nil
	}

	if strings.Count(cfg.TeamIpFormat, ".") != 3 {
		return errors.New("Invalid team ip format, expecting valid IPv4 address: " + cfg.TeamIpFormat)
	}

	return nil
}

func LoadIPs(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var ips []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ips = append(ips, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return ips, nil
}
