package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type flagIDsCCResponse map[string]map[string]map[string]map[string]any

func fetchFlagIDsCC(url, service, ipFormat string) ([]TeamTarget, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch flag IDs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var data flagIDsCCResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	serviceData, ok := data[service]
	if !ok {
		return nil, fmt.Errorf("service '%s' not found in response", service)
	}

	teamMap := make(map[int][]any)
	for teamIDStr, rounds := range serviceData {
		teamID, err := strconv.Atoi(teamIDStr)
		if err != nil || teamID <= 0 {
			continue
		}

		for _, flagIDObj := range rounds {
			teamMap[teamID] = append(teamMap[teamID], flagIDObj)
		}
	}

	if len(teamMap) == 0 {
		return nil, fmt.Errorf("no targets found for '%s'", service)
	}

	var targets []TeamTarget
	for id, flagIDs := range teamMap {
		targets = append(targets, TeamTarget{
			ID:      id,
			IP:      buildIP(ipFormat, id),
			FlagIDs: flagIDs,
		})
	}

	return targets, nil
}
