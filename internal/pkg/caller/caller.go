package caller

import (
	"runtime"
	"strings"
)

const prefixToTrim = "github.com/ShoshinNikita/budget-manager/"

// GetFormattedCaller returns formatted caller.
// It uses 'FormatCaller' function to format '*runtime.Func'
func GetFormattedCaller(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}

	details := runtime.FuncForPC(pc)
	if details == nil {
		return ""
	}

	return FormatCaller(details)
}

// FormatCaller returns formatted '*runtime.Func': something like 'some/package/path.Struct.Func'
func FormatCaller(details *runtime.Func) string {
	// funcName looks like "github.com/username/project/internal/web.Service.Func"
	funcName := details.Name()

	// Trim "github.com/username/project/"
	return strings.TrimPrefix(funcName, prefixToTrim)
}
