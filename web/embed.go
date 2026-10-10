package web

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

//go:embed all:dist
var assets embed.FS

// NewHandler serves the embedded frontend with Vue history routing support.
func NewHandler() (http.Handler, error) {
	dist, err := fs.Sub(assets, "dist")
	if err != nil {
		return nil, fmt.Errorf("加载前端资源失败: %w", err)
	}
	index, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		return nil, fmt.Errorf("加载前端首页失败，请先构建前端: %w", err)
	}
	files := http.FileServer(http.FS(dist))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "不支持的请求方法", http.StatusMethodNotAllowed)
			return
		}

		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name != "" && name != "index.html" {
			info, err := fs.Stat(dist, name)
			if err == nil && !info.IsDir() {
				files.ServeHTTP(w, r)
				return
			}
			// 不列出目录，也不将缺失的静态资源回退成 HTML。
			if err == nil || !errors.Is(err, fs.ErrNotExist) || path.Ext(name) != "" ||
				name == "assets" || strings.HasPrefix(name, "assets/") {
				http.NotFound(w, r)
				return
			}
		}

		w.Header().Set("Cache-Control", "no-cache")
		http.ServeContent(w, r, "index.html", time.Time{}, bytes.NewReader(index))
	}), nil
}
