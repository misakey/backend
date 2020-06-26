package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"github.com/pkg/errors"
)

func NewEmail(emailRenderer EmailRenderer, mailer Sender) Email {
	return Email{
		emailRenderer: emailRenderer,
		mailer:        mailer,
	}
}

type Email struct {
	emailRenderer EmailRenderer
	mailer        Sender
}

type templater interface {
	Load(name string) error
	Get(name string) (*template.Template, error)
}

// EmailNotication's content and configuration
type EmailNotification struct {
	To   string
	From string

	Subject string

	HTMLBody string
	TextBody string
}

type Renderer interface {
	NewEmail(ctx context.Context, to string, subject string, templateName string, data map[string]interface{}) (*EmailNotification, error)
}

type Sender interface {
	Send(ctx context.Context, email *EmailNotification) error
}

type EmailRenderer struct {
	templateRepo templater

	// from emails address
	mailFrom string
}

// NewEmailRenderer is mailRenderer's constructor
// It takes:
// - a templateRepo that abstract the way we get email NewTemplateFileSystem
// - a template to preload list
// - parameters about how to build emails...
func NewEmailRenderer(
	templateRepo templater,
	toLoad []string,
	mailFrom string,
) (*EmailRenderer, error) {
	renderer := &EmailRenderer{
		templateRepo: templateRepo,
		mailFrom:     mailFrom,
	}
	err := renderer.load(toLoad...)
	return renderer, err
}

// load a template inside repository
func (m *EmailRenderer) load(names ...string) error {
	var errs error
	for _, name := range names {
		err := m.templateRepo.Load(name)
		if err != nil {
			errs = errors.Wrap(errs, err.Error())
		}
	}
	if errs != nil {
		return fmt.Errorf("could not load some templates: (%s)", errs.Error())
	}
	return nil
}

// NewEmail return an new email structure filled with all necessary information to be sent.
// data must be a map[string]interface{} corresponding to template indicated by the templateN  ame string.
func (m *EmailRenderer) NewEmail(
	ctx context.Context,
	to string,
	subject string,
	templateName string,
	data map[string]interface{},
) (*EmailNotification, error) {
	email := &EmailNotification{
		To:      to,
		From:    m.mailFrom,
		Subject: subject,
	}
	// render html
	htmlBody, err := m.render(ctx, fmt.Sprintf("%s_html", templateName), data)
	if err != nil {
		return nil, err
	}
	// render text
	textBody, err := m.render(ctx, fmt.Sprintf("%s_txt", templateName), data)
	if err != nil {
		return nil, err
	}

	email.HTMLBody = string(htmlBody)
	email.TextBody = string(textBody)
	return email, nil
}

// render retrieves template from repo, executes it with given data then returns its final co  ntent
func (m *EmailRenderer) render(_ context.Context, templateName string, data map[string]interface{}) (output []byte, err error) {
	buf := &bytes.Buffer{}

	tmpl, err := m.templateRepo.Get(templateName)
	if err != nil {
		return nil, fmt.Errorf("could not get template: %v", err)
	}

	err = tmpl.Execute(buf, data)
	if err != nil {
		return nil, fmt.Errorf("could not render %s template (%v)", templateName, err)
	}

	return buf.Bytes(), nil
}
