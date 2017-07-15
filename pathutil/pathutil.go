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

func IsHidden(name string) bool {
	if strings.ContainsRune(name, os.PathSeparator) {
		name = filepath.Base(name)
	}
	return strings.HasPrefix(name, ".")
}

func IsParentHidden(name string) bool {
	if !strings.ContainsRune(name, os.PathSeparator) {
		return false
	}
	components := strings.Split(filepath.Dir(name), string(os.PathSeparator))
	for _, c := range components {
		if IsHidden(c) {
			return true
		}
	}
	return false
}
