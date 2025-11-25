package validators

import (
	"regexp"
	"unicode/utf8"
)

type TeamMemberValidator struct {
	IsValid bool
}

func NewTeamMemberValidator() *TeamMemberValidator {
	return &TeamMemberValidator{
		IsValid: true,
	}
}

func (v *TeamMemberValidator) ValidateUsername(username string) {
	if username == "" {
		v.IsValid = false
		return
	}

	// длина не больше 255 символов из-за БД
	if utf8.RuneCountInString(username) > 255 {
		v.IsValid = false
	}
}

func (v *TeamMemberValidator) ValidateId(id string) {
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

func (v *TeamMemberValidator) ValidateTeamName(teamName string) {
	if teamName == "" {
		v.IsValid = false
		return
	}

	// длина не больше 255 символов из-за БД
	if utf8.RuneCountInString(teamName) > 255 {
		v.IsValid = false
	}
}
