package logs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type LokiHook struct {
	endpoint string
	labels   map[string]string
	client   *http.Client
}

func newLokiHook(endpoint string, labels map[string]string) *LokiHook {
	return &LokiHook{
		endpoint: endpoint,
		labels:   labels,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (hook *LokiHook) Fire(entry *logrus.Entry) error {
	logData, err := entry.String()
	if err != nil {
		return fmt.Errorf("could not serialize log entry: %v", err)
	}

	lokiPayload := map[string]interface{}{
		"streams": []map[string]interface{}{
			{
				"stream": hook.labels,
				"values": [][]string{
					{fmt.Sprintf("%d", time.Now().UnixNano()), logData},
				},
			},
		},
	}

	jsonData, err := json.Marshal(lokiPayload)
	if err != nil {
		return fmt.Errorf("could not marshal JSON payload: %v", err)
	}

	req, err := http.NewRequest("POST", hook.endpoint+"/loki/api/v1/push", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("could not create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := hook.client.Do(req)
	if err != nil {
		return fmt.Errorf("could not send log to Loki: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("received non-200 response from Loki: %d", resp.StatusCode)
	}

	return nil
}

func (hook *LokiHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
