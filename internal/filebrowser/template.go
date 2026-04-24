package filebrowser

import (
	"fmt"
	"strings"
	"time"
)

func breadCrumbs(path string) []Crumb {
	crumbs := []Crumb{{Label: "~", Path: "/"}}
	parts := strings.Split(strings.Trim(path, "/"), "/")

	var acc strings.Builder

	for i, p := range parts {
		if p == "" {
			continue
		}

		acc.WriteString("/" + p)

		if i == len(parts)-1 {
			crumbs = append(crumbs, Crumb{Label: p})
		} else {
			crumbs = append(crumbs, Crumb{Label: p, Path: acc.String()})
		}
	}

	return crumbs
}

func parentOf(path string) string {
	if path == "/" {
		return "/"
	}

	trimmed := strings.TrimRight(path, "/")
	idx := strings.LastIndex(trimmed, "/")

	if idx <= 0 {
		return "/"
	}

	return trimmed[:idx]
}

func formatSize(bytes int64) string {
	if bytes == 0 {
		return "None"
	}

	units := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	f := float64(bytes)

	for f >= 1024 && i < len(units)-1 {
		f /= 1024
		i++
	}

	if i == 0 {
		return fmt.Sprintf("%d\u202f%s", int64(f), units[i])
	}

	return fmt.Sprintf("%.1f\u202f%s", f, units[i])
}

func formatDate(t time.Time) string {
	return t.Format("Jan 2, 2006 15:04")
}
