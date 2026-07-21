package submitter

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

type ENOWARSSubmitter struct {
	URL string
}

func (s *ENOWARSSubmitter) Submit(flags []Flag) ([]FlagResult, *SubmitError) {
	if len(flags) == 0 {
		return []FlagResult{}, nil
	}

	conn, err := net.DialTimeout("tcp", s.URL, 10*time.Second)
	if err != nil {
		return nil, &SubmitError{Msg: fmt.Sprintf("connect: %v", err), RateLimited: true}
	}
	defer conn.Close()

	for _, f := range flags {
		fmt.Fprintf(conn, "%s\n", f.Value)
	}

	accepted := 0
	duplicate := 0
	tooOld := 0

	scanner := bufio.NewScanner(conn)
	result := make([]FlagResult, len(flags))

	for i := 0; i < len(flags) && scanner.Scan(); i++ {
		resp := strings.TrimSpace(strings.ToLower(scanner.Text()))
		status := StatusUnknown
		switch {
		case strings.Contains(resp, "accepted"):
			accepted++
			status = StatusAccepted
		case strings.Contains(resp, "duplicate"), strings.Contains(resp, "already"):
			duplicate++
			status = StatusDuplicate
		case strings.Contains(resp, "too old"), strings.Contains(resp, "expired"):
			tooOld++
			status = StatusTooOld
		}

		result = append(result, FlagResult{
			Value:  flags[i].Value,
			Status: status,
		})
	}

	subLog.Info("Submit OK: accepted=%d duplicate=%d too_old=%d", accepted, duplicate, tooOld)
	return result, nil
}
