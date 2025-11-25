package validators

import (
	"regexp"
	"unicode/utf8"
)

// валидатор один на все хендлеры
type Validator struct {
	IsValid bool
}

func NewValidator() *Validator {
	return &Validator{
		IsValid: true,
	}
}

func (v *Validator) GetIsValid() bool {

	return v.IsValid
}

func (v *Validator) ValidatePullRequestId(id string) {
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

	// ограничение из-за БД
	if utf8.RuneCountInString(id) > 255 {
		v.IsValid = false
		return
	}
}

func (v *Validator) ValidatePullRequestName(name string) {
	if name == "" {
		v.IsValid = false
		return
	}

	if len(name) > 255 {
		v.IsValid = false
		return
	}
}

func (v *Validator) ValidateUsername(username string) {
	if username == "" {
		v.IsValid = false
		return
	}

	// длина не больше 255 символов из-за БД
	if utf8.RuneCountInString(username) > 255 {
		v.IsValid = false
	}
}

func (v *Validator) ValidateUserId(id string) {
	if id == "" {
		v.IsValid = false
		return
	}

	// Проверка формата id
	matched, _ := regexp.MatchString(`^u\d+$`, id)
	if !matched {
		v.IsValid = false
		return
	}

	// длина не больше 255 символов из-за БД
	if utf8.RuneCountInString(id) > 255 {
		v.IsValid = false
		return
	}
}

func (v *Validator) ValidateTeamName(teamName string) {
	if teamName == "" {
		v.IsValid = false
		return
	}

	// длина не больше 255 символов из-за БД
	if utf8.RuneCountInString(teamName) > 255 {
		v.IsValid = false
		return
	}
}
