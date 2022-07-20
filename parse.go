package main

import "encoding/json"

// parseLogLine tries to parse the log line as JSON and returns a generic map
// containing the result.
func parseLogLine(line []byte) (map[string]any, error) {
	lineData := map[string]any{}
	err := json.Unmarshal(line, &lineData)
	if err != nil {
		return nil, err
	}

	return lineData, nil
}
