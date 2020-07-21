package terminal

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/morikuni/aec"
)

const (
	StatusOK      = "ok"
	StatusError   = "error"
	StatusWarn    = "warn"
	StatusTimeout = "timeout"
)

var emojiStatus = map[string]string{
	StatusOK:      "\u2713",
	StatusError:   "❌",
	StatusWarn:    "⚠️",
	StatusTimeout: "⌛",
}

var textStatus = map[string]string{
	StatusOK:      " +",
	StatusError:   " !",
	StatusWarn:    " *",
	StatusTimeout: "<>",
}

var colorStatus = map[string][]aec.ANSI{
	StatusOK:    {aec.GreenF},
	StatusError: {aec.RedF},
}

// Status is used to provide an updating status to the user. The status
// usually has some animated element along with it such as a spinner.
type Status interface {
	// Update writes a new status. This should be a single line.
	Update(msg string)

	// Indicate that a step has finished, confering an ok, error, or warn upon
	// it's finishing state. If the status is not StatusOK, StatusError, or StatusWarn
	// then the status text is written directly to the output, allowing for custom
	// statuses.
	Step(status, msg string)

	// Close should be called when the live updating is complete. The
	// status will be cleared from the line.
	Close() error
}

// spinnerStatus implements Status and uses a spinner to show updates.
type spinnerStatus struct {
	mu      sync.Mutex
	spinner *spinner.Spinner
	running bool
}

var statusIcons map[string]string

const envForceEmoji = "WAYPOINT_FORCE_EMOJI"

func init() {
	if os.Getenv(envForceEmoji) != "" || strings.Contains(os.Getenv("LANG"), "UTF-8") {
		statusIcons = emojiStatus
	} else {
		statusIcons = textStatus
	}
}

func newSpinnerStatus() *spinnerStatus {
	return &spinnerStatus{
		spinner: spinner.New(
			spinner.CharSets[11],
			time.Second/6,
			spinner.WithColor("bold"),
		),
	}
}

func (s *spinnerStatus) Update(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.spinner.Suffix = " " + msg

	if !s.running {
		s.spinner.Start()
		s.running = true
	}
}

func (s *spinnerStatus) Step(status, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.spinner.Stop()
	s.running = false

	statusIcon := emojiStatus[status]
	if statusIcon == "" {
		statusIcon = status
	}

	fmt.Fprintf(color.Output, "%s %s\n", statusIcon, msg)
}

func (s *spinnerStatus) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	s.spinner.Stop()
	return nil
}

func (s *spinnerStatus) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.running = false
	s.spinner.Stop()
}

func (s *spinnerStatus) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		s.running = true
		s.spinner.Start()
	}
}
