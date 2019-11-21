// Package templates provides a store for templates which supports caching
package templates

import (
	"crypto/rand"
	"encoding/hex"
	htmlTemplate "html/template"
	"io"
	"io/ioutil"
	"sync"

	"github.com/ShoshinNikita/go-clog/v3"
)

// TemplateStore is used for serving *template.Template. It provides in-memory template caching
type TemplateStore struct {
	log *clog.Logger

	templates map[string]*template
	mux       *sync.Mutex

	getFunc func(path string) *template
}

// template is an internal wrapper for *template.Template
type template struct {
	Name string
	tpl  *htmlTemplate.Template
}

// NewTemplateStore inits new Template Store
func NewTemplateStore(log *clog.Logger, cacheTemplates bool) *TemplateStore {
	store := &TemplateStore{
		templates: make(map[string]*template),
		mux:       new(sync.Mutex),
		log:       log,
	}

	store.getFunc = store.getFromCache
	if !cacheTemplates {
		store.getFunc = store.getFromDisk
	}

	return store
}

// Get returns template with passed path. It panics, if template doesn't exist
func (t *TemplateStore) Get(path string) *htmlTemplate.Template {
	return t.getFunc(path).tpl
}

// Execute executes template with passed path. It panics, if template doesn't exist
func (t *TemplateStore) Execute(path string, w io.Writer, data interface{}) error {
	tpl := t.getFunc(path)
	return tpl.tpl.ExecuteTemplate(w, tpl.Name, data)
}

// -------------------------------------------------
// Internal methods
// -------------------------------------------------

// getFromCache tries to use cache for template. If template wasn't loaded, it calls 'getFromDisk' method
func (t *TemplateStore) getFromCache(path string) *template {
	t.mux.Lock()
	defer t.mux.Unlock()

	if tpl, ok := t.templates[path]; ok {
		// Can use cache
		t.log.Debugf("get template '%s' from cache", path)
		return tpl
	}

	// Have to load from disk
	tpl := t.getFromDisk(path)
	t.templates[path] = tpl
	return tpl
}

// getFromDisk loads template from disk
func (t *TemplateStore) getFromDisk(path string) *template {
	t.log.Debugf("load template '%s' from disk", path)

	// Don't use 'template.ParseFiles' method to support files with the same name
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic("couldn't read file: " + err.Error())
	}

	name := generateTemplateName()
	// Load from disk
	return &template{
		Name: name,
		tpl:  htmlTemplate.Must(htmlTemplate.New(name).Parse(string(data))),
	}
}

const nameLength = 8

func generateTemplateName() string {
	b := make([]byte, nameLength/2)
	rand.Read(b) //nolint:errcheck
	return hex.EncodeToString(b)
}
