package testutils

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/noqcks/solver/internal/testutils/bazel"
)

// Scanner is a convenience wrapper around a bufio.Scanner that keeps track of
// the last read line number. It also:
// - captures an associated name for the reader (typically a file name) to
//   generate positional error messages.
// - embeds a *testing.T to automatically record errors with the position it
//   corresponds to.
type Scanner struct {
	*testing.T
	*bufio.Scanner
	line int
	name string
}

func NewScanner(t *testing.T, r io.Reader, name string, line int) *Scanner {
	bufioScanner := bufio.NewScanner(r)
	// We use a large max-token-size to account for lines in the output that far
	// exceed the default bufio Scanner token size.
	bufioScanner.Buffer(make([]byte, 100), 10*bufio.MaxScanTokenSize)
	// TODO(irfansharif): Detect if we're running under bazel, and if so, strip
	// out the sandbox path prefix.
	if bazel.BuiltWithBazel() {
		name = strings.TrimPrefix(name, bazel.ScratchDirectory(t))
	}
	return &Scanner{
		T:       t,
		Scanner: bufioScanner,
		line:    line,
		name:    name,
	}
}

// Scan is a thin wrapper around the bufio Scanner's interface.
func (s *Scanner) Scan() bool {
	ok := s.Scanner.Scan()
	if ok {
		s.line++
	}
	return ok
}

// Fatal is thin wrapper around testing.T's interface.
func (s *Scanner) Fatal(args ...interface{}) {
	s.T.Fatalf("%s: %s", s.pos(), fmt.Sprint(args...))
}

// Fatalf is thin wrapper around testing.T's interface.
func (s *Scanner) Fatalf(format string, args ...interface{}) {
	s.T.Fatalf("%s: %s", s.pos(), fmt.Sprintf(format, args...))
}

// Error is thin wrapper around testing.T's interface.
func (s *Scanner) Error(args ...interface{}) {
	s.T.Errorf("%s: %s", s.pos(), fmt.Sprint(args...))
}

// Errorf is thin wrapper around testing.T's interface.
func (s *Scanner) Errorf(format string, args ...interface{}) {
	s.T.Errorf("%s: %s", s.pos(), fmt.Sprintf(format, args...))
}

// Log is thin wrapper around testing.T's interface.
func (s *Scanner) Log(args ...interface{}) {
	s.T.Logf("%s: %s", s.pos(), fmt.Sprint(args...))
}

// Logf is thin wrapper around testing.T's interface.
func (s *Scanner) Logf(format string, args ...interface{}) {
	s.T.Logf("%s: %s", s.pos(), fmt.Sprint(args...))
}

// pos is a file:line prefix for the input file, suitable for inclusion in logs
// and error messages.
func (s *Scanner) pos() string {
	return fmt.Sprintf("%s:%d", s.name, s.line)
}
