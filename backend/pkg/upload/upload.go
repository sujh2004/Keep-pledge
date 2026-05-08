package upload

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

func SaveOptional(c *gin.Context, field string, dir string, publicPath string, maxSizeMB int64) (string, error) {
	file, err := c.FormFile(field)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return "", nil
		}
		return "", err
	}
	return save(c, file, dir, publicPath, maxSizeMB)
}

func save(c *gin.Context, file *multipart.FileHeader, dir string, publicPath string, maxSizeMB int64) (string, error) {
	if maxSizeMB <= 0 {
		maxSizeMB = 5
	}
	if file.Size > maxSizeMB*1024*1024 {
		return "", fmt.Errorf("图片不能超过 %dMB", maxSizeMB)
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExtensions[ext] {
		return "", errors.New("仅支持 jpg/png/webp 图片")
	}
	datePath := time.Now().Format("20060102")
	targetDir := filepath.Join(dir, datePath)
	name := randomName() + ext
	target := filepath.Join(targetDir, name)
	if err := c.SaveUploadedFile(file, target); err != nil {
		return "", err
	}
	return strings.TrimRight(publicPath, "/") + "/" + datePath + "/" + name, nil
}

func randomName() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}
