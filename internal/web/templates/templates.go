// Package templates provides a store for templates which supports caching
package templates

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	htmlTemplate "html/template"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
)

// TemplateStore is used for serving *template.Template. It provides in-memory template caching
type TemplateStore struct {
	log logrus.FieldLogger

	templates map[string]*template
	mux       *sync.Mutex

	getFunc func(path string, log logrus.FieldLogger) *template
}

// template is an internal wrapper for *template.Template
type template struct {
	Name string
	tpl  *htmlTemplate.Template
}

// NewTemplateStore inits new Template Store
func NewTemplateStore(log logrus.FieldLogger, cacheTemplates bool) *TemplateStore {
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
func (t *TemplateStore) Get(ctx context.Context, path string) *htmlTemplate.Template {
	log := request_id.FromContextToLogger(ctx, t.log)
	log = log.WithField("path", path)

	return t.getFunc(path, log).tpl
}

// Execute executes template with passed path. It panics, if template doesn't exist
// Execute checks for errors before writing into w: it executes template into
// temporary buffer and copies data if everything is fine
func (t *TemplateStore) Execute(ctx context.Context, path string,
	w io.Writer, data interface{}) error {

	log := request_id.FromContextToLogger(ctx, t.log)
	log = log.WithField("path", path)

	tpl := t.getFunc(path, log)
	buff := bytes.NewBuffer(nil)

	now := time.Now()
	err := tpl.tpl.ExecuteTemplate(buff, tpl.Name, data)
	log = log.WithField("time", time.Since(now))
	if err != nil {
		log.WithError(err).Error("couldn't execute template")
		return err
	}

	log.Debug("execute template successfully")
	_, err = io.Copy(w, buff)
	return err
}

// -------------------------------------------------
// Internal methods
// -------------------------------------------------

// getFromCache tries to use cache for template. If template wasn't loaded, it calls 'getFromDisk' method
func (t *TemplateStore) getFromCache(path string, log logrus.FieldLogger) *template {
	t.mux.Lock()
	defer t.mux.Unlock()

	if tpl, ok := t.templates[path]; ok {
		// Can use cache
		log.Debug("get template from cache")
		return tpl
	}

	// Have to load from disk
	tpl := t.getFromDisk(path, log)
	t.templates[path] = tpl
	return tpl
}

// getFromDisk loads template from disk
func (t *TemplateStore) getFromDisk(path string, log logrus.FieldLogger) *template {
	log.Debug("load template from disk")

	// Don't use 'template.ParseFiles' method to support files with the same name
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.WithError(err).Panic("couldn't read file")
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
