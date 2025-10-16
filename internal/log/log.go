// Copyright (c) 2025 Steve Taranto <staranto@gmail.com>.
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/apex/log"
)

// InitLogger sets up Apex with a custom handler and a log level from the
// TFCTL_LOG env variable.
func InitLogger() {
	level := strings.ToUpper(os.Getenv("TFCTL_LOG"))
	if level == "" {
		level = "ERROR"
	}
	log.SetHandler(&CustomHandler{})
	log.SetLevelFromString(level)
}

// CustomHandler formats log messages and writes to stdout
type CustomHandler struct{}

// HandleLog implements the log.Handler interface
func (h *CustomHandler) HandleLog(e *log.Entry) error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	level := strings.ToUpper(e.Level.String())
	message := e.Message
	fmt.Fprintf(os.Stdout, "%s %.1s %s\n", timestamp, level, message)
	return nil
}
