package targets

import (
	"encoding/json"
	"regexp"
)

type Email struct{}

type EmailTarget struct {
	Email
	Addresses []*EmailAddress
}

type EmailAddress struct {
	To      string
	ReplyTo string
}

const emailTargetTemplate = "email-edit.html"

var (
	emailRegex = regexp.MustCompile(`^[\w\.\-\+\_]+@[\w\.\-]+\.[a-zA-Z]+$`)
)

func init() {
	addTargetType(Email{})
}

func (e Email) Id() TargetTypeId {
	return 1
}

func (e Email) Name() string {
	return "Email"
}

func (e Email) Load(target string) (Target, error) {
	var et EmailTarget
	err := json.Unmarshal([]byte(target), &et)
	if err != nil {
		return nil, err
	}
	return et, nil
}

func (et Email) Templates() map[string]string {
	return map[string]string{
		"edit": "email-edit.html",
		"view": "email-view.html",
	}
}

func (et Email) Scripts() map[string][]string {
	return map[string][]string{
		"edit": []string{
			"email.js",
		},
	}
}

func (et EmailTarget) Validate() (errs []string) {
	for _, e := range et.Addresses {
		if !emailRegex.MatchString(e.To) {
			errs = append(errs, "An invalid email to was provided.")
			break
		}
	}
	for _, e := range et.Addresses {
		if e.ReplyTo == "" {
			continue
		}
		if !emailRegex.MatchString(e.ReplyTo) {
			errs = append(errs, "An invalid reply to was provided.")
			break
		}
	}
	return
}

func (et EmailTarget) TargetType() TargetType {
	return Email{}
}
