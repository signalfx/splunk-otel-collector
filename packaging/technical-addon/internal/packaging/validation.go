package packaging

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"strings"
)

func Sha256Sum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(strings.TrimSpace(string(data))))
	return hex.EncodeToString(sum[:]), nil
}
