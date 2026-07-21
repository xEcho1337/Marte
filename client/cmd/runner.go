package cmd

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed templates/template_head.py
var templateHead string

//go:embed templates/template_tail.py
var templateTail string

//go:embed templates/sample_exploit.py
var samplePy string

type TeamTarget struct {
	ID      int
	IP      string
	FlagIDs []any
}

func fetchFlagIDs(url, service, format, ipFormat string) ([]TeamTarget, error) {
	if format == "forcad" {
		return fetchFlagIDsForcAD(url, service, ipFormat)
	}
	if format == "enowars" {
		return fetchFlagIDsENOWARS(url, service, ipFormat)
	}
	return fetchFlagIDsCC(url, service, ipFormat)
}

func buildIP(format string, teamID int) string {
	return strings.Replace(format, "{}", strconv.Itoa(teamID), 1)
}

type runInput struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	FlagIDs []any  `json:"flag_ids"`
}

func runPythonExploit(tempFile, host string, port int, flagIDs []any, showStdout bool, flagPattern string) []string {
	input := runInput{
		Host:    host,
		Port:    port,
		FlagIDs: flagIDs,
	}
	inputJSON, _ := json.Marshal(input)

	python := findPython()
	cmd := exec.Command(python, tempFile)
	cmd.Stdin = bytes.NewReader(inputJSON)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Error("Error while executing: " + err.Error())
	}

	if showStdout {
		os.Stdout.Write(stdout.Bytes())
		if stderr.Len() > 0 {
			log.Error("Python traceback: " + stderr.String())
		}
	}

	cleaned := stdout.String()
	cleaned = regexp.MustCompile(`b'(.*?)'`).ReplaceAllString(cleaned, "$1")
	cleaned = regexp.MustCompile(`b"(.*?)"`).ReplaceAllString(cleaned, "$1")

	var re *regexp.Regexp
	if flagPattern == "" {
		re = regexp.MustCompile(`(?i)flag\{[^}]+\}`)
	} else {
		re = regexp.MustCompile(flagPattern)
	}
	matches := re.FindAllString(cleaned, -1)

	seen := make(map[string]struct{})
	var flags []string

	for _, f := range matches {
		if _, ok := seen[f]; ok {
			continue
		}

		seen[f] = struct{}{}
		flags = append(flags, f)
	}

	return flags
}

func findPython() string {
	for _, name := range []string{"python3", "python"} {
		if _, err := exec.LookPath(name); err == nil {
			return name
		}
	}
	return "python3"
}

type runMode int

const (
	runModeAttack runMode = iota
	runModeTest
	runModeDebug
)

func createTempFile(exploitPath string) (string, error) {
	exploitContent, err := os.ReadFile(exploitPath)
	if err != nil {
		return "", fmt.Errorf("reading exploit: %w", err)
	}

	combined := templateHead + "\n" + string(exploitContent) + "\n" + templateTail

	f, err := os.CreateTemp("", "marte_*.py")
	if err != nil {
		return "", err
	}

	if _, err := f.WriteString(combined); err != nil {
		f.Close()
		return "", err
	}
	f.Close()

	return f.Name(), nil
}

func executeRound(cfg *Config, exploitPath, serviceName, explicitHost string, mode runMode) {
	servicePort, ok := cfg.Services[serviceName]

	if !ok {
		log.Error("Service not found: " + serviceName + ". Run 'marte pull'")
		return
	}

	var targets []TeamTarget

	if explicitHost != "" {
		var flagIDs []any
		if fullTargets, err := fetchFlagIDs(cfg.UrlFlagIDs, serviceName, cfg.FlagIDFormat, cfg.TeamIPFormat); err == nil {
			for _, t := range fullTargets {
				if t.IP == explicitHost {
					flagIDs = t.FlagIDs
					break
				}
			}
		}
		targets = []TeamTarget{{ID: 0, IP: explicitHost, FlagIDs: flagIDs}}
	} else {
		log.Info("Attacking...")
		switch mode {
		case runModeAttack:
			if cfg.TeamCount <= 0 {
				log.Error("Team count not configured")
				return
			}
			flagIDMap := fetchFlagIDMap(cfg, serviceName)
			for id := 1; id <= cfg.TeamCount; id++ {
				if id == cfg.TeamID {
					continue
				}
				targets = append(targets, TeamTarget{
					ID:      id,
					IP:      buildIP(cfg.TeamIPFormat, id),
					FlagIDs: flagIDMap[id],
				})
			}

		case runModeTest, runModeDebug:
			log.Info("Debug mode...")
			fullTargets, err := fetchFlagIDs(cfg.UrlFlagIDs, serviceName, cfg.FlagIDFormat, cfg.TeamIPFormat)
			if err != nil {
				log.Error("Flag IDs: " + err.Error())
				return
			}
			for _, t := range fullTargets {
				if t.ID == cfg.NOPTeamID {
					targets = []TeamTarget{t}
					break
				}
			}
			if len(targets) == 0 {
				log.Errorf("NOP team %d not found in flag IDs", cfg.NOPTeamID)
				return
			}
		}
	}

	if len(targets) == 0 {
		log.Info("No valid targets")
		return
	}

	tempFile, err := createTempFile(exploitPath)
	if err != nil {
		log.Error("Template error: " + err.Error())
		return
	}
	defer os.Remove(tempFile)

	var wg sync.WaitGroup
	for _, t := range targets {
		wg.Add(1)
		go func(target TeamTarget) {
			defer wg.Done()

			ip := target.IP
			if ip == "" {
				ip = buildIP(cfg.TeamIPFormat, target.ID)
			}
			log.Info("Team %d (%s):%d", target.ID, ip, servicePort)

			debug := mode == runModeDebug
			flags := runPythonExploit(tempFile, ip, servicePort, target.FlagIDs, debug, cfg.FlagRegex)

			if len(flags) == 0 {
				return
			}

			log.Info("Team %d: %d flags found", target.ID, len(flags))

			if mode == runModeDebug {
				for _, f := range flags {
					log.Info("  > %s", f)
				}
				return
			}

			addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.SubmitterPort))
			conn, err := net.Dial("tcp", addr)

			if err != nil {
				log.Errorf("Submit failed: %v", err)
				return
			}

			defer conn.Close()

			var submitting []flagData

			for _, f := range flags {
				submitting = append(submitting, flagData{
					TargetTeam: target.ID, Value: f, Service: serviceName,
				})
			}

			data := submitData{
				Token:     cfg.AccessToken,
				Submitter: cfg.Username,
				Flags:     submitting,
			}

			if err := writeSubmitData(conn, data); err != nil {
				log.Errorf("Write submit: %v", err)
				return
			}
		}(t)
	}

	wg.Wait()
}

func fetchFlagIDMap(cfg *Config, service string) map[int][]any {
	targets, err := fetchFlagIDs(cfg.UrlFlagIDs, service, cfg.FlagIDFormat, cfg.TeamIPFormat)
	if err != nil {
		return nil
	}
	m := make(map[int][]any, len(targets))
	for _, t := range targets {
		m[t.ID] = t.FlagIDs
	}
	return m
}

func runExploitLoop(cfg *Config, exploitPath, serviceName string, mode runMode) {
	for {
		roundStart := time.Now()
		modeStr := "ATTACK"
		if mode == runModeDebug {
			modeStr = "DEBUG"
		} else if mode == runModeTest {
			modeStr = "TEST"
		}
		fmt.Print("\n")
		log.Info("Round %s (%s) %s", serviceName, roundStart.Format("15:04:05"), modeStr)

		executeRound(cfg, exploitPath, serviceName, "", mode)

		elapsed := time.Since(roundStart)
		sleepDuration := time.Duration(cfg.RoundTime)*time.Second - elapsed
		if sleepDuration > 0 {
			log.Info("Round in %v, attesa %v...",
				elapsed.Round(time.Millisecond),
				sleepDuration.Round(time.Second))
			time.Sleep(sleepDuration)
		} else {
			log.Info("Round in %v (> %ds), continuo",
				elapsed.Round(time.Millisecond), cfg.RoundTime)
		}
	}
}

func runExploitOnce(cfg *Config, exploitPath, serviceName, explicitHost string, mode runMode) {
	executeRound(cfg, exploitPath, serviceName, explicitHost, mode)
}
