package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/hako/durafmt"
	"github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"
)

type alertResponse struct {
	Status string        `json:"status"`
	Alerts []types.Alert `json:"data,omitempty"`
}

func listAlerts(logger log.Logger, alertmanagerURL string) ([]types.Alert, error) {
	resp, err := httpGetRetry(logger, alertmanagerURL+"/api/v1/alerts")
	if err != nil {
		return nil, err
	}

	var alertResponse alertResponse
	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	if err := dec.Decode(&alertResponse); err != nil {
		return nil, err
	}

	return alertResponse.Alerts, err
}

// AlertMessage converts an alert to a message string
func AlertMessage(a types.Alert) string {
	var status, duration string
	switch a.Status() {
	case model.AlertFiring:
		status = fmt.Sprintf("🔥 *%s* 🔥", strings.ToUpper(string(a.Status())))
		duration = fmt.Sprintf("*Started*: %s ago", durafmt.Parse(time.Since(a.StartsAt)))
	case model.AlertResolved:
		status = fmt.Sprintf("*%s*", strings.ToUpper(string(a.Status())))
		duration = fmt.Sprintf(
			"*Ended*: %s ago\n*Duration*: %s",
			durafmt.Parse(time.Since(a.EndsAt)),
			durafmt.Parse(a.EndsAt.Sub(a.StartsAt)),
		)
	}

	return fmt.Sprintf(
		"%s\n*%s* (%s)\n%s\n%s\n",
		status,
		a.Labels["alertname"],
		a.Annotations["summary"],
		a.Annotations["description"],
		duration,
	)
}
