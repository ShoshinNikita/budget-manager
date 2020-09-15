package templates

import (
	"context"
	htmlTemplate "html/template"
	"io"

	"github.com/sirupsen/logrus"

	reqid "github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
)

// TemplateDiskExecutor reads templates from disk every execution
type TemplateDiskExecutor struct {
	log logrus.FieldLogger
}

func NewTemplateDiskExecutor(log logrus.FieldLogger) *TemplateDiskExecutor {
	return &TemplateDiskExecutor{
		log: log,
	}
}

func (ex *TemplateDiskExecutor) Get(ctx context.Context, template Template) *htmlTemplate.Template {
	log := reqid.FromContextToLogger(ctx, ex.log)
	log = log.WithField("path", template.Path)

	log.Debug("load template from disk")
	tpl, err := loadTemplateWithDeps(template)
	if err != nil {
		ex.log.WithError(err).Panic("couldn't load template")
	}

	return tpl
}

func (ex *TemplateDiskExecutor) Execute(ctx context.Context, template Template, w io.Writer, data interface{}) error {
	tpl := ex.Get(ctx, template)
	return executeTemplate(ex.log, tpl, w, data)
}
