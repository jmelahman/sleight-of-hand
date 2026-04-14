package dispatch

import (
	"sort"

	"github.com/jmelahman/sleight-of-hand/internal/passthrough"
	"github.com/jmelahman/sleight-of-hand/tools/gh"
)

var registry = map[string]func([]string) int{
	"gh": gh.Run,
}

// Run dispatches to the registered handler for toolName, or falls back
// to a pure passthrough if the tool is not registered.
func Run(toolName string, args []string) int {
	handler, ok := registry[toolName]
	if !ok {
		return passthrough.Exec(toolName, args)
	}
	return handler(args)
}

// RegisteredTools returns the sorted list of tool names in the registry.
func RegisteredTools() []string {
	tools := make([]string, 0, len(registry))
	for name := range registry {
		tools = append(tools, name)
	}
	sort.Strings(tools)
	return tools
}

// IsRegistered reports whether toolName is in the registry.
func IsRegistered(toolName string) bool {
	_, ok := registry[toolName]
	return ok
}
