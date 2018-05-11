package pathutil

import (
	"os"
	"path/filepath"
	"strings"
)

func Depth(name string) int {
	name = filepath.Clean(name)
	return strings.Count(name, string(os.PathSeparator))
}

func ContainsHidden(path string) bool {
	names := strings.Split(path, string(os.PathSeparator))
	for _, name := range names {
		if strings.HasPrefix(name, ".") {
			return true
		}
	}
	return false
}
