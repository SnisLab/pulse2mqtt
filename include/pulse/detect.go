package pulse

import (
	"fmt"
	"strings"
	"sync"

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
	if probe(baseURL + ModernDataPath() + query) {
		state.Lock()
		state.mode = ModeModern
		state.Unlock()
		log.Info("Detected modern Pulse API")
		return nil
	}
	if probe(baseURL + LegacyDataPath() + query) {
		state.Lock()
		state.mode = ModeLegacy
		state.Unlock()
		log.Info("Detected legacy Pulse API")
		return nil
	}
	return fmt.Errorf("could not reach Pulse data endpoint (tried node_data.json and data.json)")
}

func probe(url string) bool {
	resp, err := Get(url)
	if err != nil {
		log.Error("Pulse endpoint probe failed:", url, err)
		return false
	}
	if resp.StatusCode != 200 {
		log.Error("Pulse endpoint probe returned:", url, resp.Status)
		return false
	}
	return len(resp.Body) > 0
}

func CurrentMode() Mode {
	state.RLock()
	defer state.RUnlock()
	return state.mode
}
