package greynoise

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// doGet executes req, reads the body, checks for HTTP errors, and JSON-decodes into out.
// The caller is responsible for setting request headers before calling.
func doGet(ctx context.Context, hc *http.Client, req *http.Request, out any) error {
	req = req.WithContext(ctx)
	resp, err := hc.Do(req)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("json decode: %w", err)
	}
	return nil
}
