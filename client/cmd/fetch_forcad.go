package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type flagIDsForcADResponse map[string]map[string][]string

func fetchFlagIDsForcAD(url, service, ipFormat string) ([]TeamTarget, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch flag IDs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var data flagIDsForcADResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	serviceData, ok := data[service]
	if !ok {
		return nil, fmt.Errorf("service '%s' not found in response", service)
	}

	var targets []TeamTarget
	for key, raw := range serviceData {
		teamID, ip, ok := resolveTeam(key, ipFormat)
		if !ok {
			continue
		}

		flagIDs := extractFlagIDs(raw)
		if len(flagIDs) == 0 {
			continue
		}

		targets = append(targets, TeamTarget{
			ID:      teamID,
			IP:      ip,
			FlagIDs: flagIDs,
		})
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no targets found for '%s'", service)
	}

	return targets, nil
}

func resolveTeam(key, ipFormat string) (teamID int, ip string, ok bool) {
	if id, err := strconv.Atoi(key); err == nil && id > 0 {
		return id, buildIP(ipFormat, id), true
	}
	if id, ip, ok := extractFromIP(key, ipFormat); ok {
		return id, ip, true
	}
	return 0, "", false
}

func extractFromIP(ip, ipFormat string) (teamID int, resolved string, ok bool) {
	ipParts := strings.Split(ip, ".")
	fmtParts := strings.Split(ipFormat, ".")

	if len(ipParts) != 4 || len(fmtParts) != 4 {
		return 0, "", false
	}

	for i, part := range fmtParts {
		if part == "{}" {
			id, err := strconv.Atoi(ipParts[i])
			if err != nil || id <= 0 {
				return 0, "", false
			}
			return id, ip, true
		}
	}
	if id, err := strconv.Atoi(ipParts[3]); err == nil && id > 0 {
		return id, ip, true
	}
	return 0, "", false
}

func extractFlagIDs(items []string) []any {
	var out []any
	for _, item := range items {
		if item == "" {
			continue
		}

		if item[0] == '{' {
			var parsed any
			if json.Unmarshal([]byte(item), &parsed) == nil {
				out = append(out, parsed)
			} else {
				out = append(out, item)
			}
		} else {
			out = append(out, parsePairString(item))
		}
	}
	return out
}

func parsePairString(s string) map[string]any {
	m := make(map[string]any)
	parts := strings.Split(s, ",")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(kv) == 2 {
			m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	if len(m) == 0 {
		m["_raw"] = s
	}
	return m
}
