package validators

import (
	"regexp"
	"sync"
	"unicode/utf8"
)

type IValidator interface {
	GetIsValid() bool
	SetIsValidFalse()
	ValidatePullRequestId(id string)
	ValidatePullRequestName(name string)
	ValidateUsername(username string)
	ValidateUserId(id string)
	ValidateTeamName(teamName string)
}

// валидатор один на все хендлеры
type Validator struct {
	IsValid bool
	mtx     sync.RWMutex
}

func NewValidator() *Validator {
	return &Validator{
		IsValid: true,
		mtx:     sync.RWMutex{},
	}
}

func (v *Validator) SetIsValidFalse() {
	v.mtx.Lock()
	defer v.mtx.Unlock()
	v.IsValid = false
}

func (v *Validator) GetIsValid() bool {
	v.mtx.RLock()
	defer v.mtx.RUnlock()
	return v.IsValid
}

func (v *Validator) ValidatePullRequestId(id string) {
	if id == "" {
		v.SetIsValidFalse()
		return
	}

	// Проверка формата prid
	matched, _ := regexp.MatchString(`^pr-\d+$`, id)
	if !matched {
		v.SetIsValidFalse()
		return
	}

	// ограничение из-за БД
	if utf8.RuneCountInString(id) > 255 {
		v.SetIsValidFalse()
		return
	}
}

func (v *Validator) ValidatePullRequestName(name string) {
	if name == "" {
		v.SetIsValidFalse()
		return
	}

	if len(name) > 255 {
		v.SetIsValidFalse()
		return
	}
}

func (v *Validator) ValidateUsername(username string) {
	if username == "" {
		v.SetIsValidFalse()
		return
	}

	// длина не больше 255 символов из-за БД
	if utf8.RuneCountInString(username) > 255 {
		v.SetIsValidFalse()
	}
}

func (v *Validator) ValidateUserId(id string) {
	if id == "" {
		v.SetIsValidFalse()
		return
	}

	// Проверка формата id
	matched, _ := regexp.MatchString(`^u\d+$`, id)
	if !matched {
		v.SetIsValidFalse()
		return
	}

	// длина не больше 255 символов из-за БД
	if utf8.RuneCountInString(id) > 255 {
		v.SetIsValidFalse()
		return
	}
}

func (v *Validator) ValidateTeamName(teamName string) {
	if teamName == "" {
		v.SetIsValidFalse()
		return
	}

	// длина не больше 255 символов из-за БД
	if utf8.RuneCountInString(teamName) > 255 {
		v.SetIsValidFalse()
	}
}
