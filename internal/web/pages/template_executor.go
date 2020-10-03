package pages

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	reqid "github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
)

type templateExecutor struct {
	cacheTemplates bool
	log            logrus.FieldLogger

	mu  sync.Mutex
	tpl *template.Template
}

func newTemplateExecutor(log logrus.FieldLogger, cacheTemplates bool) *templateExecutor {
	return &templateExecutor{
		log:            log,
		cacheTemplates: cacheTemplates,
	}
}

func (e *templateExecutor) Execute(ctx context.Context, w io.Writer, name string, data interface{}) error {
	log := reqid.FromContextToLogger(ctx, e.log)

	tpl, err := e.loadTemplates()
	if err != nil {
		return errors.Wrap(err, "couldn't load templates")
	}

	tpl = tpl.Lookup(name)
	if tpl == nil {
		return errors.Errorf("no template with name '%s'", name)
	}

	if err := executeTemplate(log, tpl, w, data); err != nil {
		return errors.Wrap(err, "couldn't execute template")
	}

	return nil
}

const templatesDir = "./templates"

// loadTemplates loads all templates from file or returns them from cache according to 'cacheTemplates'
func (e *templateExecutor) loadTemplates() (_ *template.Template, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.cacheTemplates && e.tpl != nil {
		return e.tpl, nil
	}

	var templates []string
	err = filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		templates = append(templates, path)
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get template filenames")
	}

	if e.tpl, err = template.ParseFiles(templates...); err != nil {
		return nil, err
	}

	return e.tpl, nil
}

// executeTemplate executes passed template. It checks for errors before writing into w: it executes
// template into temporary buffer and copies data if everything is fine
func executeTemplate(log logrus.FieldLogger, tpl *template.Template, w io.Writer, data interface{}) error {
	buff := bytes.NewBuffer(nil)

	now := time.Now()
	if err := tpl.Execute(buff, data); err != nil {
		return err
	}
	log.WithField("time", time.Since(now)).Debug("template was successfully executed")

	_, err := io.Copy(w, buff)
	return err
}
