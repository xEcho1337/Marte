package submitter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type HTTPSSubmitter struct {
	Client    *http.Client
	URL       string
	TeamToken string
	Name      string
}

func (s *HTTPSSubmitter) Submit(flags []Flag) ([]FlagResult, *SubmitError) {
	if len(flags) == 0 {
		return []FlagResult{}, nil
	}

	flagList := make([]string, len(flags))
	for i, f := range flags {
		flagList[i] = f.Value
	}

	body, err := json.Marshal(flagList)
	if err != nil {
		return nil, &SubmitError{Msg: fmt.Sprintf("marshal flags: %v", err)}
	}

	req, err := http.NewRequest(http.MethodPut, s.URL, bytes.NewReader(body))
	if err != nil {
		return nil, &SubmitError{Msg: fmt.Sprintf("create request: %v", err)}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Team-Token", s.TeamToken)

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, &SubmitError{Msg: fmt.Sprintf("http request: %v", err), RateLimited: true}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, &SubmitError{Msg: "rate limited", RateLimited: true}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, &SubmitError{Msg: fmt.Sprintf("unexpected status %s: %s", resp.Status, string(bodyBytes))}
	}

	bodyBytes, _ := io.ReadAll(resp.Body)

	var result []FlagResult
	var submitResponse []map[string]string

	if err := json.Unmarshal(bodyBytes, &submitResponse); err == nil {
		tooOld := 0
		alreadyStolen := 0
		accepted := 0
		for _, flagReply := range submitResponse {
			status := StatusUnknown
			msg := strings.ToLower(flagReply["msg"])

			if strings.Contains(msg, "too old") {
				tooOld++
				status = StatusTooOld
			} else if strings.Contains(msg, "already claimed") || strings.Contains(msg, "already stolen") {
				alreadyStolen++
				status = StatusDuplicate
			} else if strings.Contains(msg, "accepted") {
				accepted++
				status = StatusAccepted
			} else {
				subLog.Info("Unknown flag reply: %s", flagReply)
			}

			result = append(result, FlagResult{
				Value:  flagReply["flag"],
				Status: status,
			})
		}
		subLog.Info("Submit OK: accepted=%d too_old=%d already_stolen=%d", accepted, tooOld, alreadyStolen)
	} else {
		subLog.Info("Submit OK (raw): %s", string(bodyBytes))
	}

	return result, nil
}
