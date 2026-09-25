package pulse

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	log "github.com/DjSni/go-log"

	"pulse2mqtt/include/settings"
)

type Mode string

const (
	ModeLegacy Mode = "legacy"
	ModeModern Mode = "modern"
)

var state struct {
	sync.RWMutex
	mode Mode
}

// Initialize selects the Pulse API mode once during application startup.
func Initialize() error {
	configured := strings.ToLower(strings.TrimSpace(settings.Load.Service.Pulse.Version))
	if configured == string(ModeLegacy) || configured == string(ModeModern) {
		state.Lock()
		state.mode = Mode(configured)
		state.Unlock()
		log.Info("Using configured Pulse API mode:", configured)
		return nil
	}

	baseURL := "http://" + settings.Load.Service.Pulse.IP
	query := "?node_id=" + fmt.Sprint(settings.Load.Service.Pulse.Node)
	client := &http.Client{Timeout: 10 * time.Second}

	if probe(client, baseURL+"/node_data.json"+query) {
		state.Lock()
		state.mode = ModeModern
		state.Unlock()
		log.Info("Detected modern Pulse API")
		return nil
	}
	if probe(client, baseURL+"/data.json"+query) {
		state.Lock()
		state.mode = ModeLegacy
		state.Unlock()
		log.Info("Detected legacy Pulse API")
		return nil
	}
	return fmt.Errorf("could not reach Pulse data endpoint (tried node_data.json and data.json)")
}

func probe(client *http.Client, url string) bool {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Debug("Can not create Pulse endpoint probe:", err)
		return false
	}
	req.SetBasicAuth(settings.Load.Service.Pulse.User, settings.Load.Service.Pulse.Password)
	resp, err := client.Do(req)
	if err != nil {
		log.Debug("Pulse endpoint probe failed:", url, err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Debug("Pulse endpoint probe returned:", url, resp.Status)
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Debug("Can not read Pulse endpoint probe response:", err)
		return false
	}
	return len(body) > 0
}

func DataPath() string {
	state.RLock()
	defer state.RUnlock()
	if state.mode == ModeModern {
		return "/node_data.json"
	}
	return "/data.json"
}

func MetricsPath() string {
	state.RLock()
	defer state.RUnlock()
	if state.mode == ModeModern {
		return "/node_metrics.json"
	}
	return "/metrics.json"
}
