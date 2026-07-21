package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var httpClient = &http.Client{Timeout: 3 * time.Second}

type Config struct {
	Host          string         `json:"host"`
	Port          int            `json:"port"`
	Username      string         `json:"username"`
	RoundTime     int            `json:"round_time"`
	AccessToken   string         `json:"access_token"`
	ConfigSum     string         `json:"config_sum"`
	Services      map[string]int `json:"services"`
	SubmitterPort int            `json:"submitter_port"`
	NOPTeamID     int            `json:"nop_team_id"`
	TeamID        int            `json:"team_id"`
	TeamIPFormat  string         `json:"team_ip_format"`
	UrlFlagIDs    string         `json:"url_flag_ids"`
	TeamToken     string         `json:"team_token"`
	FlagIDFormat  string         `json:"flag_id_format"`
	TeamCount     int            `json:"team_count"`
	FlagRegex     string         `json:"flag_regex"`
}

func (c *Config) ComputeSum() string {
	data, _ := json.Marshal(c.Services)
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config not found, run 'marte init'")
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config")
	}
	return &cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create and activate an environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := ".marte"
		if err := os.MkdirAll(path, 0755); err != nil {
			log.Error("Cannot create .marte directory")
			return nil
		}
		cfg := Config{
			Host:          "127.0.0.1",
			Port:          14100,
			RoundTime:     120,
			SubmitterPort: 14101,
			NOPTeamID:     0,
			TeamID:        1,
			TeamIPFormat:  "10.10.{}.1",
			UrlFlagIDs:    "http://10.10.0.1:8081/flagIds",
			FlagIDFormat:  "cyber_challenge",
			TeamCount:     10,
			FlagRegex:     `(?i)flag\{[^}]+\}`,
			Services:      make(map[string]int),
		}
		if err := SaveConfig(path+"/config.json", &cfg); err != nil {
			log.Error("Cannot save config")
			return nil
		}
		log.Info("Environment created in .marte/")
		return nil
	},
}

var hostCmd = &cobra.Command{
	Use:   "host <ip> <port>",
	Short: "Set the backend IP and port for the current environment",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := strconv.ParseInt(args[1], 10, 0)

		if err != nil {
			log.Error("Invalid port: " + args[1])
			return nil
		}

		cfg, err := LoadConfig(".marte/config.json")

		if err != nil {
			log.Error(err.Error())
			return nil
		}

		cfg.Host = args[0]
		cfg.Port = int(port)

		if err := SaveConfig(".marte/config.json", cfg); err != nil {
			log.Error("Cannot save config")
			return nil
		}

		log.Info("Host: %s:%d", cfg.Host, cfg.Port)
		return nil
	},
}

var loginCmd = &cobra.Command{
	Use:   "login <username> <password>",
	Short: "Authenticate on the backend and save the token",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := LoadConfig(".marte/config.json")

		if err != nil {
			log.Error(err.Error())
			return nil
		}

		url := fmt.Sprintf("http://%s:%d/api/login", cfg.Host, cfg.Port)
		body, _ := json.Marshal(map[string]string{
			"username": args[0],
			"password": args[1],
		})

		resp, err := httpClient.Post(url, "application/json", bytes.NewReader(body))

		if err != nil {
			log.Error("Cannot connect to backend: " + err.Error())
			return nil
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			msg, _ := io.ReadAll(resp.Body)
			log.Error("Login failed: " + string(msg))
			return nil
		}

		var result struct {
			Authorization string `json:"authorization"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Error("Invalid server response")
			return nil
		}

		cfg.Username = args[0]
		cfg.AccessToken = result.Authorization

		if err := SaveConfig(".marte/config.json", cfg); err != nil {
			log.Error("Cannot save token")
			return nil
		}

		log.Info("Logged in, token saved")
		return nil
	},
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Update services and configuration from the backend",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := LoadConfig(".marte/config.json")

		if err != nil {
			log.Error(err.Error())
			return nil
		}

		if cfg.AccessToken == "" {
			log.Error("Missing token, run 'marte login' first")
			return nil
		}

		url := fmt.Sprintf("http://%s:%d/api/get_services", cfg.Host, cfg.Port)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)

		resp, err := httpClient.Do(req)

		if err != nil {
			log.Error("Cannot connect to backend")
			return nil
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			msg, _ := io.ReadAll(resp.Body)
			log.Error("Pull services failed: " + string(msg))
			return nil
		}

		var result struct {
			Services []struct {
				Name string `json:"name"`
				Port int    `json:"port"`
			} `json:"services"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Error("Invalid server response")
			return nil
		}

		services := make(map[string]int, len(result.Services))
		for _, s := range result.Services {
			services[s.Name] = s.Port
		}
		cfg.Services = services

		farmURL := fmt.Sprintf("http://%s:%d/api/farm", cfg.Host, cfg.Port)
		farmReq, _ := http.NewRequest("GET", farmURL, nil)
		farmReq.Header.Set("Authorization", "Bearer "+cfg.AccessToken)

		farmResp, err := httpClient.Do(farmReq)
		if err == nil && farmResp.StatusCode == http.StatusOK {
			defer farmResp.Body.Close()
			var farm struct {
				SubmitterPort int    `json:"submitter_port"`
				NOPTeamID     int    `json:"nop_team_id"`
				TeamID        int    `json:"team_id"`
				TeamIPFormat  string `json:"team_ip_format"`
				UrlFlagIDs    string `json:"url_flag_ids"`
				TeamToken     string `json:"team_token"`
				FlagIDFormat  string `json:"flag_id_format"`
				TeamCount     int    `json:"team_count"`
				FlagRegex     string `json:"flag_regex"`
			}
			if json.NewDecoder(farmResp.Body).Decode(&farm) == nil {
				cfg.SubmitterPort = farm.SubmitterPort
				cfg.NOPTeamID = farm.NOPTeamID
				cfg.TeamID = farm.TeamID
				cfg.TeamIPFormat = farm.TeamIPFormat
				cfg.UrlFlagIDs = farm.UrlFlagIDs
				cfg.TeamToken = farm.TeamToken
				cfg.FlagIDFormat = farm.FlagIDFormat
				cfg.TeamCount = farm.TeamCount
				cfg.FlagRegex = farm.FlagRegex
			}
		}

		cfg.ConfigSum = cfg.ComputeSum()

		if err := SaveConfig(".marte/config.json", cfg); err != nil {
			log.Error("Cannot save config")
			return nil
		}

		log.Info("Downloaded %d services", len(services))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd, hostCmd, loginCmd, pullCmd)
}
