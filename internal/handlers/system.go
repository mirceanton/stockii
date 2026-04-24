package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/disintegration/imaging"
	"github.com/mirceanton/stockii/internal/db"
	"github.com/mirceanton/stockii/internal/models"
)

type StorageStats struct {
	DBBytes          int64   `json:"db_bytes"`
	DBBytesHuman     string  `json:"db_bytes_human"`
	ImagesBytes      int64   `json:"images_bytes"`
	ImagesBytesHuman string  `json:"images_bytes_human"`
	ImagesCount      int     `json:"images_count"`
	VolumeTotalBytes uint64  `json:"volume_total_bytes"`
	VolumeTotalHuman string  `json:"volume_total_human"`
	VolumeFreeBytes  uint64  `json:"volume_free_bytes"`
	VolumeFreeHuman  string  `json:"volume_free_human"`
	DBUsedPct        float64 `json:"db_used_pct"`
	AssetsUsedPct    float64 `json:"assets_used_pct"`
	VolumeUsedPct    float64 `json:"volume_used_pct"`
	VolumeBarClass   string  `json:"volume_bar_class"`
}

func computeStorageStats() StorageStats {
	var stats StorageStats

	if info, err := os.Stat(GetConfigPath()); err == nil {
		stats.DBBytes = info.Size()
	}
	stats.DBBytesHuman = formatBytes(stats.DBBytes)

	imagesDir := filepath.Join(GetDataPath(), "images")
	filepath.WalkDir(imagesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if info, e := d.Info(); e == nil {
			stats.ImagesBytes += info.Size()
			stats.ImagesCount++
		}
		return nil
	})
	stats.ImagesBytesHuman = formatBytes(stats.ImagesBytes)

	var statfs syscall.Statfs_t
	if err := syscall.Statfs(GetDataPath(), &statfs); err == nil {
		stats.VolumeTotalBytes = statfs.Blocks * uint64(statfs.Bsize)
		stats.VolumeFreeBytes = statfs.Bfree * uint64(statfs.Bsize)
	}
	stats.VolumeTotalHuman = formatBytes(int64(stats.VolumeTotalBytes))
	stats.VolumeFreeHuman = formatBytes(int64(stats.VolumeFreeBytes))

	if stats.VolumeTotalBytes > 0 {
		total := float64(stats.VolumeTotalBytes)
		stats.DBUsedPct = clamp100(float64(stats.DBBytes) / total * 100)
		stats.AssetsUsedPct = clamp100(float64(stats.DBBytes+stats.ImagesBytes) / total * 100)
		stats.VolumeUsedPct = clamp100(float64(stats.VolumeTotalBytes-stats.VolumeFreeBytes) / total * 100)
	}

	switch {
	case stats.VolumeUsedPct > 90:
		stats.VolumeBarClass = "bg-danger"
	case stats.VolumeUsedPct > 75:
		stats.VolumeBarClass = "bg-warning"
	default:
		stats.VolumeBarClass = "bg-secondary"
	}

	return stats
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	stats := computeStorageStats()
	render(w, "status.html", map[string]interface{}{
		"Page":  "status",
		"Stats": stats,
	})
}

func StorageStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats := computeStorageStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func RecompressImagesHandler(w http.ResponseWriter, r *http.Request) {
	imagesDir := filepath.Join(GetDataPath(), "images")

	entries, err := os.ReadDir(imagesDir)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<div class="alert alert-danger mb-0">Failed to read images directory: %s</div>`, err)
		return
	}

	processed, errCount := 0, 0
	var errMsgs []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		oldName := entry.Name()
		oldPath := filepath.Join(imagesDir, oldName)

		// Decode first so we can check for alpha before imaging.Fit normalises the type.
		src, err := imaging.Open(oldPath, imaging.AutoOrientation(true))
		if err != nil {
			errCount++
			errMsgs = append(errMsgs, fmt.Sprintf("%s: %v", oldName, err))
			continue
		}

		hasAlpha := imageHasAlpha(src)
		img := imaging.Fit(src, 1500, 1500, imaging.Lanczos)

		stem := strings.TrimSuffix(oldName, filepath.Ext(oldName))
		newExt := ".jpg"
		if hasAlpha {
			newExt = ".png"
		}
		newName := stem + newExt
		newPath := filepath.Join(imagesDir, newName)
		tmpPath := newPath + ".tmp"

		if err := encodeImage(img, tmpPath, hasAlpha); err != nil {
			errCount++
			errMsgs = append(errMsgs, fmt.Sprintf("%s: save failed: %v", newName, err))
			os.Remove(tmpPath)
			continue
		}

		if err := os.Rename(tmpPath, newPath); err != nil {
			errCount++
			errMsgs = append(errMsgs, fmt.Sprintf("%s: rename failed: %v", newName, err))
			os.Remove(tmpPath)
			continue
		}

		if oldName != newName {
			db.DB.Model(&models.Product{}).Where("image_path = ?", oldName).Update("image_path", newName)
			os.Remove(oldPath)
		}

		processed++
	}

	w.Header().Set("Content-Type", "text/html")
	if errCount == 0 {
		fmt.Fprintf(w, `<div class="alert alert-success mb-0"><i class="bi bi-check-circle me-1"></i>Done — %d image(s) recompressed. <a href="/status" class="alert-link">Refresh stats</a></div>`, processed)
	} else {
		fmt.Fprintf(w, `<div class="alert alert-warning mb-0"><i class="bi bi-exclamation-triangle me-1"></i>%d recompressed, %d error(s). <a href="/status" class="alert-link">Refresh stats</a></div>`, processed, errCount)
	}
}

func clamp100(v float64) float64 {
	if v > 100 {
		return 100
	}
	return v
}

func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
