package caller

import (
	"runtime"
	"strings"
)

const (
	packageNameToTrim  = "github.com/ShoshinNikita/budget-manager/"
	internalPathToTrim = "internal/"
)

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
	funcName = strings.TrimPrefix(funcName, packageNameToTrim)
	// Trim "internal/"
	funcName = strings.TrimPrefix(funcName, internalPathToTrim)

	return funcName
}
