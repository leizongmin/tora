package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolveFilePath(root string, url string) (string, error) {
	if url == "" || url == "/" {
		return root, nil
	}
	p, err := filepath.Abs(filepath.Join(root, url[1:]))
	if err != nil {
		return p, err
	}
	if strings.Index(p, root+string(os.PathSeparator)) >= 0 {
		return p, nil
	}
	return p, fmt.Errorf("cannot access to %s", url)
}
