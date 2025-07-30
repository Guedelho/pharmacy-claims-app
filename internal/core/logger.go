package core

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type Logger struct {
	logDir string
}

func NewLogger(logDir string) *Logger {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create log directory: %v", err)
	}

	return &Logger{logDir: logDir}
}

func (l *Logger) LogEvent(eventType string, payload map[string]interface{}) {
	event := map[string]interface{}{
		"timestamp":  fmt.Sprintf("%d", uuid.New().ID()),
		"event_type": eventType,
		"payload":    payload,
	}

	filename := fmt.Sprintf("%s-%s.json", eventType, uuid.New().String())
	filepath := filepath.Join(l.logDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		log.Printf("Failed to create log file %s: %v", filepath, err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(event); err != nil {
		log.Printf("Failed to encode event to file %s: %v", filepath, err)
	}
}
