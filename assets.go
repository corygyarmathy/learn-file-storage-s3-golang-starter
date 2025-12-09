package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"mime"
	"os"
	"path/filepath"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetKey(mediaType string) string {
	base := make([]byte, 32)
	_, err := rand.Read(base)
	if err != nil {
		panic("failed to generate random bytes")
	}
	id := base64.RawURLEncoding.EncodeToString(base)

	ext := mediaTypeToExt(mediaType)
	return fmt.Sprintf("%s%s", id, ext)
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func (cfg apiConfig) getObjectURL(key string) string {
	return fmt.Sprintf("https://%s/%s", cfg.s3CfDistribution, key)
}

func mediaTypeToExt(mediaType string) string {
	exts, _ := mime.ExtensionsByType(mediaType)
	ext := ""
	if len(exts) > 0 {
		ext = exts[0]
	}
	return ext
}

func validateImageMediaType(mediaType string) bool {
	if mediaType != "image/jpeg" && mediaType != "image/png" {
		return false
	}
	return true
}

func validateVideoMediaType(mediaType string) bool {
	return mediaType == "video/mp4"
}
