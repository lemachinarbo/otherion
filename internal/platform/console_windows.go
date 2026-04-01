//go:build windows

package platform

import "golang.org/x/sys/windows"

// AttachConsole allocates a console window for log output.
// On Windows, GUI applications have no console by default (built with -H windowsgui).
// Call this before logging initialization so os.Stderr output is visible.
func AttachConsole() {
	windows.AllocConsole()
}
