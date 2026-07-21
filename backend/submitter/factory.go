package submitter

import (
	"net/http"
	"strings"
	"time"
)

func createSubmitter(submitterType, urlFlagSubmit, teamToken string) Submitter {
	switch submitterType {
	case "cyber_challenge", "forcad":
		return &HTTPSSubmitter{
			Client: &http.Client{
				Timeout: 10 * time.Second,
			},
			URL:       urlFlagSubmit,
			TeamToken: teamToken,
			Name:      submitterType,
		}
	case "enowars":
		return &ENOWARSSubmitter{
			URL: strings.TrimPrefix(urlFlagSubmit, ""),
		}
	case "faust":
		return &FaustSubmitter{
			URL: strings.TrimPrefix(urlFlagSubmit, ""),
		}
	default:
		return nil
	}
}
