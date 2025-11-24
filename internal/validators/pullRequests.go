package validators

import (
	"regexp"
	"unicode/utf8"
)

type PullRequestValidator struct {
	IsValid bool
}

func NewPullRequestValidator() *PullRequestValidator {
	return &PullRequestValidator{
		IsValid: true,
	}
}

func (v *PullRequestValidator) ValidatePullRequestId(id string) {
	if id == "" {
		v.IsValid = false
		return
	}

	// Проверка формата prid
	matched, _ := regexp.MatchString(`^pr-\d+$`, id)
	if !matched {
		v.IsValid = false
		return
	}

	//ограничение из-за БД
	if utf8.RuneCountInString(id) > 255 {
		v.IsValid = false
		return
	}
}

func (v *PullRequestValidator) ValidatePullRequestName(name string) {
	if name == "" {
		v.IsValid = false
		return
	}

	if len(name) > 255 {
		v.IsValid = false
		return
	}
}

func (v *PullRequestValidator) ValidateAuthorId(authorId string) {
	if authorId == "" {
		v.IsValid = false
		return
	}

	// Проверка формата user_id
	matched, _ := regexp.MatchString(`^u\d+$`, authorId)
	if !matched {
		v.IsValid = false
		return
	}
}
