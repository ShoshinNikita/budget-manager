package pages

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	reqid "github.com/ShoshinNikita/budget-manager/internal/pkg/request_id"
	"github.com/ShoshinNikita/budget-manager/templates"
)

type templateExecutor struct {
	cacheTemplates bool
	fs             fs.ReadDirFS
	log            logrus.FieldLogger
	commonFuncs    template.FuncMap

	mu  sync.Mutex
	tpl *template.Template
}

func newTemplateExecutor(log logrus.FieldLogger, cacheTemplates bool, commonFuncs template.FuncMap) *templateExecutor {
	return &templateExecutor{
		fs:             templates.New(cacheTemplates),
		log:            log,
		cacheTemplates: cacheTemplates,
		commonFuncs:    commonFuncs,
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

// loadTemplates loads all templates from file or returns them from cache according to 'cacheTemplates'
func (e *templateExecutor) loadTemplates() (_ *template.Template, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.cacheTemplates && e.tpl != nil {
		return e.tpl, nil
	}

	patterns, err := extractAllTemplatePaths(e.fs)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get template filenames")
	}

	e.tpl, err = template.New("base").Funcs(e.getCommonFuncs()).ParseFS(e.fs, patterns...)
	if err != nil {
		return nil, err
	}

	return e.tpl, nil
}

func (e *templateExecutor) getCommonFuncs() template.FuncMap {
	res := make(template.FuncMap, len(e.commonFuncs))
	for k, v := range e.commonFuncs {
		res[k] = v
	}
	return res
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

func extractAllTemplatePaths(fs fs.ReadDirFS) ([]string, error) {
	const maxDepth = 25

	var walk func(root string, depth int) ([]string, error)
	walk = func(root string, depth int) (paths []string, err error) {
		if depth >= maxDepth {
			return nil, errors.Errorf("max dir depth is reached: %d", maxDepth)
		}

		entries, err := fs.ReadDir(root)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't read dir")
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				if isTemplate(entry.Name()) {
					paths = append(paths, filepath.Join(root, entry.Name()))
				}
				continue
			}

			nestedPaths, err := walk(filepath.Join(root, entry.Name()), depth+1)
			if err != nil {
				return nil, err
			}
			paths = append(paths, nestedPaths...)
		}
		return paths, nil
	}

	return walk(".", 0)
}

func isTemplate(name string) bool {
	ext := filepath.Ext(name)

	return ext == ".html"
}
