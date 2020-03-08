package blocklist

import (
	"encoding/json"
	"fmt"
	"time"
)

type HistoryEntry struct {
	ResourceType string    `json:"type"`
	Source       string    `json:"source"`
	Host         string    `json:"host"`
	Timestamp    time.Time `json:"timestamp"`
	Block        bool      `json:"block"`
}

func (he *HistoryEntry) String() string {
	return fmt.Sprintf("%v: %v", he.Source, he.Host)
}

func (he *HistoryEntry) JSONBytes() []byte {
	bytes, _ := json.Marshal(*he)
	return bytes
}
