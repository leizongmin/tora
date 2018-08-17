package file

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
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
	if strings.Index(p, root+string(os.PathSeparator)) > 0 {
		return p, nil
	}
	return p, fmt.Errorf("cannot access to %s", url)
}

func getFileMd5(filePath string) (string, error) {
	var returnMD5String string
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil
}
