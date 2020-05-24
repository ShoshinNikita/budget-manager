package templates

import (
	"context"
	htmlTemplate "html/template"
	"io"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
)

// TemplateCacheExecutor caches templates after reading them from the disk
type TemplateCacheExecutor struct {
	log logrus.FieldLogger

	templates map[string]*htmlTemplate.Template
	mu        sync.RWMutex
}

var _ TemplateExecutor = (*TemplateCacheExecutor)(nil)

func NewTemplateCacheExecutor(log logrus.FieldLogger) *TemplateCacheExecutor {
	return &TemplateCacheExecutor{
		log: log,
		//
		templates: make(map[string]*htmlTemplate.Template),
	}
}

func (ex *TemplateCacheExecutor) Get(ctx context.Context, template Template) *htmlTemplate.Template {
	log := request_id.FromContextToLogger(ctx, ex.log)
	log = log.WithField("path", template.Path)

	ex.mu.RLock()
	tpl, ok := ex.templates[template.Path]
	ex.mu.RUnlock()
	if ok {
		log.Debug("use template from cache")
		return tpl
	}

	// Have to load template from disk
	log.Debug("load template from disk")
	tpl, err := loadTemplateWithDeps(template)
	if err != nil {
		ex.log.WithError(err).Panic("couldn't load template")
	}

	ex.mu.Lock()
	ex.templates[template.Path] = tpl
	ex.mu.Unlock()

	return tpl
}

func (ex *TemplateCacheExecutor) Execute(ctx context.Context, template Template, w io.Writer, data interface{}) error {
	tpl := ex.Get(ctx, template)
	return executeTemplate(ex.log, tpl, w, data)
}
