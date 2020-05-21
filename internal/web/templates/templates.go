// Package templates provides a store for templates which supports caching
package templates

import (
	"bytes"
	"context"
	htmlTemplate "html/template"
	"io"
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type TemplateExecutor interface {
	// Get returns template with passed path. It should panic if template doesn't exist
	Get(ctx context.Context, t Template) *htmlTemplate.Template
	// Execute executes template. It should panic if template doesn't exist
	Execute(ctx context.Context, t Template, w io.Writer, data interface{}) error
}

type Template struct {
	Path string
	Deps []string
}

// ----------------------------------------------------
// Helpers
// ----------------------------------------------------

// loadTemplateWithDeps load base template and associates all deps with it
func loadTemplateWithDeps(template Template) (*htmlTemplate.Template, error) {
	// Read base template
	baseTpl, err := loadTemplateFromDisk(template.Path)
	if err != nil {
		return nil, err
	}

	// Load deps
	for _, dep := range template.Deps {
		depTpl, err := loadTemplateFromDisk(dep)
		if err != nil {
			return nil, err
		}
		if _, err = baseTpl.AddParseTree(dep, depTpl.Tree.Copy()); err != nil {
			return nil, errors.Wrap(err, "couldn't add parsed tree")
		}
	}

	return baseTpl, nil
}

// loadTemplateFromDisk reads file and parses template
func loadTemplateFromDisk(path string) (*htmlTemplate.Template, error) {
	tpl := htmlTemplate.New(path)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't read file")
	}
	if _, err := tpl.Parse(string(data)); err != nil {
		return nil, errors.Wrap(err, "couldn't parse template")
	}

	return tpl, nil
}

// executeTemplate executes passed template. It checks for errors before writing into w: it executes
// template into temporary buffer and copies data if everything is fine
func executeTemplate(log logrus.FieldLogger, tpl *htmlTemplate.Template, w io.Writer, data interface{}) error {
	buff := bytes.NewBuffer(nil)

	now := time.Now()
	if err := tpl.Execute(buff, data); err != nil {
		return err
	}
	log.WithField("time", time.Since(now)).Debug("template was successfully executed")

	_, err := io.Copy(w, buff)
	return err
}
