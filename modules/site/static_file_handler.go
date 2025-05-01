package site

import (
	"bytes"
	"embed"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// StaticFileHandler 处理静态文件的结构体
// 添加对嵌入文件的支持

type StaticFileHandler struct {
	cache          sync.Map
	cacheMutex     sync.Mutex
	cacheTTL       time.Duration
	embedFS        embed.FS
	BaseRoot       string
	useEmbed       bool
	forceIndexHTML bool
}

type StaticFileHandlerConfig struct {
	TTL            time.Duration
	EmbedFS        embed.FS
	BaseRoot       string
	UseEmbed       bool
	ForceIndexHTML bool
}

// NewStaticFileHandler 创建一个新的静态文件处理器
func NewStaticFileHandler(config StaticFileHandlerConfig) *StaticFileHandler {
	return &StaticFileHandler{
		cacheTTL:       config.TTL,
		embedFS:        config.EmbedFS,
		BaseRoot:       config.BaseRoot,
		useEmbed:       config.UseEmbed,
		forceIndexHTML: config.ForceIndexHTML,
	}
}

// ServeStaticFile 提供静态文件服务
func (h *StaticFileHandler) ServeStaticFile(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		r.URL.Path = "/index.html"
	}
	requestedPath := filepath.Join(h.BaseRoot, r.URL.Path)
	// 检查缓存
	if cachedContent, ok := h.cache.Load(requestedPath); ok {
		if content, valid := cachedContent.(cachedFile); valid && time.Since(content.timestamp) < h.cacheTTL {
			http.ServeContent(w, r, requestedPath, time.Now(), bytes.NewReader(content.data))
			return
		}
		h.cache.Delete(requestedPath) // 删除过期缓存
	}

	var content []byte
	if h.forceIndexHTML {
		if _, err := os.Stat(requestedPath); os.IsNotExist(err) {
			requestedPath = filepath.Join(h.BaseRoot, "index.html")
		}
	}

	if h.useEmbed {
		// 从嵌入文件系统读取文件

		file, err := h.embedFS.Open(requestedPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()

		content, err = io.ReadAll(file)
		if err != nil {
			http.NotFound(w, r)
			return
		}
	} else {
		// 从文件系统读取文件
		file, err := os.Open(requestedPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()

		content, err = io.ReadAll(file)
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}

	// 缓存文件内容
	h.cache.Store(requestedPath, cachedFile{
		data:      content,
		timestamp: time.Now(),
	})

	http.ServeContent(w, r, requestedPath, time.Now(), bytes.NewReader(content))
}

// 清理过期缓存
type cachedFile struct {
	data      []byte
	timestamp time.Time
}

func (h *StaticFileHandler) StartCacheCleaner() {
	go func() {
		for {
			time.Sleep(h.cacheTTL)
			h.cacheMutex.Lock()
			h.cache.Range(func(key, value interface{}) bool {
				if content, valid := value.(cachedFile); valid && time.Since(content.timestamp) >= h.cacheTTL {
					h.cache.Delete(key)
				}
				return true
			})
			h.cacheMutex.Unlock()
		}
	}()
}
