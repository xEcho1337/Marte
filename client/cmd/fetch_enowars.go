package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type flagIDsENOWARSResponse struct {
	Services map[string]map[string]map[string]map[string][]string `json:"services"`
}

func fetchFlagIDsENOWARS(url, service, ipFormat string) ([]TeamTarget, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch flag IDs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var data flagIDsENOWARSResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	serviceData, ok := data.Services[service]
	if !ok {
		return nil, fmt.Errorf("service '%s' not found in response", service)
	}

	teamMap := make(map[string][]any)
	for ip, rounds := range serviceData {
		for roundStr, flagTypes := range rounds {
			roundNum, _ := strconv.Atoi(roundStr)
			for fType, values := range flagTypes {
				for _, val := range values {
					entry := map[string]any{
						"round": roundNum,
						"type":  fType,
						"value": val,
					}
					teamMap[ip] = append(teamMap[ip], entry)
				}
			}
		}
	}

	if len(teamMap) == 0 {
		return nil, fmt.Errorf("no targets found for '%s'", service)
	}

	var targets []TeamTarget
	for ip, flagIDs := range teamMap {
		id, _, ok := resolveTeam(ip, ipFormat)
		if !ok {
			continue
		}
		targets = append(targets, TeamTarget{
			ID:      id,
			IP:      ip,
			FlagIDs: flagIDs,
		})
	}

	return targets, nil
}
