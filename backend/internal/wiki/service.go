// Package wiki — раздача содержимого docs/wiki/ru/ через HTTP.
// План 19 (game-wiki), MVP.
//
// Отдаёт .md файлы как plaintext + извлечённый frontmatter в JSON.
// Frontmatter — простой key: value между двумя `---` в начале файла.
//
// Защита от path traversal: разрешены только имена вида
// [a-zA-Z0-9_-]+ без '..', без абсолютных путей.
package wiki

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ErrNotFound — страница не найдена.
var ErrNotFound = errors.New("wiki: page not found")

// ErrBadPath — попытка обойти sandbox.
var ErrBadPath = errors.New("wiki: invalid path")

// Service читает файлы из корня wiki (обычно docs/wiki/ru/).
type Service struct {
	root string
}

// NewService — root должен быть абсолютным или относительным от CWD
// сервера.
func NewService(root string) *Service {
	return &Service{root: root}
}

// Page — одна страница вики.
type Page struct {
	Path        string            `json:"path"`        // "buildings/metal_mine"
	Frontmatter map[string]string `json:"frontmatter"`
	Markdown    string            `json:"markdown"`    // содержимое после frontmatter
}

// Category — один пункт в боковом меню.
type Category struct {
	Key   string `json:"key"`   // "buildings"
	Title string `json:"title"` // "Здания"
	Order int    `json:"order"`
}

// List возвращает каталоги первого уровня (категории).
func (s *Service) List() ([]Category, error) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		return nil, fmt.Errorf("wiki list: %w", err)
	}
	var cats []Category
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		key := e.Name()
		if !safeName(key) {
			continue
		}
		p, err := s.Get(key + "/index")
		if err != nil {
			continue
		}
		title := p.Frontmatter["title"]
		if title == "" {
			title = key
		}
		order := parseOrderInt(p.Frontmatter["order"])
		cats = append(cats, Category{Key: key, Title: title, Order: order})
	}
	sort.Slice(cats, func(i, j int) bool {
		if cats[i].Order != cats[j].Order {
			return cats[i].Order < cats[j].Order
		}
		return cats[i].Key < cats[j].Key
	})
	return cats, nil
}

// ListCategory возвращает список страниц в категории.
func (s *Service) ListCategory(cat string) ([]Page, error) {
	if !safeName(cat) {
		return nil, ErrBadPath
	}
	dir := filepath.Join(s.root, cat)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("wiki ls category: %w", err)
	}
	var pages []Page
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		slug := strings.TrimSuffix(e.Name(), ".md")
		p, err := s.Get(cat + "/" + slug)
		if err != nil {
			continue
		}
		p.Markdown = "" // в листинге тело не отдаём
		pages = append(pages, *p)
	}
	sort.Slice(pages, func(i, j int) bool {
		oi := parseOrderInt(pages[i].Frontmatter["order"])
		oj := parseOrderInt(pages[j].Frontmatter["order"])
		if oi != oj {
			return oi < oj
		}
		return pages[i].Path < pages[j].Path
	})
	return pages, nil
}

// Get читает страницу по пути вида "buildings/metal_mine" (без .md).
// "index" допустимо как slug — тогда читается index.md категории.
// Если путь = "index" — читается корневой docs/wiki/ru/index.md.
func (s *Service) Get(path string) (*Page, error) {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil, ErrBadPath
	}
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		return nil, ErrBadPath
	}
	for _, p := range parts {
		if !safeName(p) {
			return nil, ErrBadPath
		}
	}
	rel := filepath.Join(parts...) + ".md"
	full := filepath.Join(s.root, rel)
	data, err := os.ReadFile(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("wiki read: %w", err)
	}
	fm, md := splitFrontmatter(string(data))
	return &Page{Path: path, Frontmatter: fm, Markdown: md}, nil
}

// --- helpers ---

// safeName — разрешены имена из latin-цифр, дефис и подчёркивание.
// Не допускаем точек, слешей, '..', пустых строк.
func safeName(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return false
		}
	}
	return true
}

func splitFrontmatter(data string) (map[string]string, string) {
	fm := map[string]string{}
	if !strings.HasPrefix(data, "---\n") && !strings.HasPrefix(data, "---\r\n") {
		return fm, data
	}
	// Отрезаем первый '---'
	rest := strings.TrimPrefix(data, "---\n")
	rest = strings.TrimPrefix(rest, "---\r\n")
	// Ищем закрывающий '---'
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return fm, data
	}
	head := rest[:idx]
	body := rest[idx+4:] // '\n---' = 4
	body = strings.TrimPrefix(body, "\n")
	body = strings.TrimPrefix(body, "\r\n")
	for _, ln := range strings.Split(head, "\n") {
		ln = strings.TrimRight(ln, "\r")
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		colon := strings.IndexByte(ln, ':')
		if colon < 0 {
			continue
		}
		k := strings.TrimSpace(ln[:colon])
		v := strings.TrimSpace(ln[colon+1:])
		fm[k] = v
	}
	return fm, body
}

func parseOrderInt(s string) int {
	if s == "" {
		return 1000
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 1000
		}
		n = n*10 + int(r-'0')
	}
	return n
}
