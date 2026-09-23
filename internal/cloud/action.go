package cloud

import (
	"fmt"
	"os"
	"strings"
)

// WriteActionOutputs appends acknowledged publication outputs without permitting
// remote response values to inject additional workflow output records.
func WriteActionOutputs(path string, response *RunIngestResponse) error {
	if path == "" {
		path = os.Getenv("GITHUB_OUTPUT")
	}
	if path == "" {
		return nil
	}
	if strings.ContainsAny(response.RunID+response.URL, "\r\n\x00") {
		return fmt.Errorf("invalid control character in cloud action output")
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	_, writeErr := fmt.Fprintf(file, "cloud-run-id=%s\ncloud-run-url=%s\nis-regression=%t\n", response.RunID, response.URL, response.IsRegression)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}
