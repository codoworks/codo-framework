package errors

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
)

func init() {
	// Respect NO_COLOR environment variable (https://no-color.org/)
	if os.Getenv("NO_COLOR") != "" {
		color.NoColor = true
	}
}

// RenderCLI renders an error for command-line output with colors and formatting
func RenderCLI(err error) {
	if err == nil {
		return
	}

	// Map to framework error
	mappedErr := MapError(err)

	// Color scheme based on severity
	var (
		headerColor *color.Color
		borderColor *color.Color
	)

	labelColor := color.New(color.FgCyan, color.Bold)
	valueColor := color.New(color.FgWhite)
	dimColor := color.New(color.FgHiBlack)

	switch {
	case mappedErr.HTTPStatus >= 500:
		headerColor = color.New(color.FgRed, color.Bold)
		borderColor = color.New(color.FgRed)
	case mappedErr.HTTPStatus >= 400:
		headerColor = color.New(color.FgYellow, color.Bold)
		borderColor = color.New(color.FgYellow)
	default:
		headerColor = color.New(color.FgBlue, color.Bold)
		borderColor = color.New(color.FgBlue)
	}

	border := strings.Repeat("━", 70)

	// Header
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, borderColor.Sprint(border))
	fmt.Fprintf(os.Stderr, "%s %s\n", headerColor.Sprint("ERROR:"), headerColor.Sprint(mappedErr.Message))
	fmt.Fprintln(os.Stderr, borderColor.Sprint(border))

	// Error code
	fmt.Fprintf(os.Stderr, "%s %s\n", labelColor.Sprint("Code:"), valueColor.Sprint(mappedErr.Code))

	// Phase
	if mappedErr.Phase != "" {
		fmt.Fprintf(os.Stderr, "%s %s\n", labelColor.Sprint("Phase:"), valueColor.Sprint(mappedErr.Phase))
	}

	// Location
	if mappedErr.Caller != nil {
		location := fmt.Sprintf("%s:%d", mappedErr.Caller.File, mappedErr.Caller.Line)
		fmt.Fprintf(os.Stderr, "%s %s\n", labelColor.Sprint("Location:"), valueColor.Sprint(location))
		fmt.Fprintf(os.Stderr, "%s %s\n", labelColor.Sprint("Function:"), valueColor.Sprint(mappedErr.Caller.Function))
	}

	// Timestamp with timezone
	fmt.Fprintf(os.Stderr, "%s %s\n", labelColor.Sprint("Time:"), valueColor.Sprint(mappedErr.Timestamp.Format(time.RFC3339)))

	// Details (sorted for consistent output)
	if len(mappedErr.Details) > 0 {
		fmt.Fprintf(os.Stderr, "\n%s\n", labelColor.Sprint("Details:"))

		// Sort keys for consistent output
		keys := make([]string, 0, len(mappedErr.Details))
		for k := range mappedErr.Details {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", dimColor.Sprint(k), valueColor.Sprint(mappedErr.Details[k]))
		}
	}

	// Error chain
	if mappedErr.Cause != nil {
		fmt.Fprintf(os.Stderr, "\n%s\n", labelColor.Sprint("Error Chain:"))
		renderErrorChain(mappedErr, "  ", dimColor, valueColor)
	}

	// Stack trace (if present)
	if len(mappedErr.StackTrace) > 0 {
		fmt.Fprintf(os.Stderr, "\n%s\n", labelColor.Sprint("Stack Trace:"))
		for i, frame := range mappedErr.StackTrace {
			if i >= 10 {
				fmt.Fprintf(os.Stderr, "  %s\n", dimColor.Sprint("... (truncated)"))
				break
			}
			fmt.Fprintf(os.Stderr, "  %s %s:%d\n", dimColor.Sprint("→"), valueColor.Sprint(frame.File), frame.Line)
			fmt.Fprintf(os.Stderr, "    %s\n", dimColor.Sprint(frame.Function))
		}
	}

	// Footer
	fmt.Fprintln(os.Stderr, borderColor.Sprint(border))
	fmt.Fprintln(os.Stderr)
}

func renderErrorChain(err *Error, indent string, dimColor, valueColor *color.Color) {
	current := err.Cause
	depth := 0
	maxDepth := 5

	for current != nil && depth < maxDepth {
		fmt.Fprintf(os.Stderr, "%s%s %s\n", indent, dimColor.Sprint("→"), valueColor.Sprint(current.Error()))

		// If it's our error type, show additional context
		if e, ok := current.(*Error); ok {
			if e.Caller != nil {
				fmt.Fprintf(os.Stderr, "%s  %s %s:%d\n", indent, dimColor.Sprint("at"), dimColor.Sprint(e.Caller.File), e.Caller.Line)
			}
			current = e.Cause
		} else {
			// Try to unwrap
			type unwrapper interface {
				Unwrap() error
			}
			if u, ok := current.(unwrapper); ok {
				current = u.Unwrap()
			} else {
				current = nil
			}
		}

		depth++
	}

	if current != nil {
		fmt.Fprintf(os.Stderr, "%s%s\n", indent, dimColor.Sprint("... (more errors in chain)"))
	}
}
