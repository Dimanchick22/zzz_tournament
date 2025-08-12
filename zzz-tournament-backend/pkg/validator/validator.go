// pkg/validator/validator.go
package validator

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,50}$`)
	passwordRegex = regexp.MustCompile(`^.{6,}$`)
	roomNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-_]{3,255}$`)
	phoneRegex    = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	urlRegex      = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
)

type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Value   interface{} `json:"value,omitempty"`
}

// Error реализует интерфейс error для ValidationError
func (v *ValidationError) Error() string {
	return v.Message
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var messages []string
	for _, err := range v {
		messages = append(messages, err.Field+": "+err.Message)
	}
	return strings.Join(messages, ", ")
}

func (v ValidationErrors) HasErrors() bool {
	return len(v) > 0
}

func (v ValidationErrors) GetField(field string) *ValidationError {
	for _, err := range v {
		if err.Field == field {
			return &err
		}
	}
	return nil
}

// Email validation
func ValidateEmail(email string) *ValidationError {
	if email == "" {
		return &ValidationError{
			Field:   "email",
			Message: "Email is required",
			Code:    "REQUIRED",
		}
	}

	if len(email) > 254 {
		return &ValidationError{
			Field:   "email",
			Message: "Email is too long (maximum 254 characters)",
			Code:    "TOO_LONG",
			Value:   email,
		}
	}

	if !emailRegex.MatchString(email) {
		return &ValidationError{
			Field:   "email",
			Message: "Invalid email format",
			Code:    "INVALID_FORMAT",
			Value:   email,
		}
	}

	return nil
}

// Username validation
func ValidateUsername(username string) *ValidationError {
	if username == "" {
		return &ValidationError{
			Field:   "username",
			Message: "Username is required",
			Code:    "REQUIRED",
		}
	}

	if len(username) < 3 {
		return &ValidationError{
			Field:   "username",
			Message: "Username must be at least 3 characters long",
			Code:    "TOO_SHORT",
			Value:   username,
		}
	}

	if len(username) > 50 {
		return &ValidationError{
			Field:   "username",
			Message: "Username must be at most 50 characters long",
			Code:    "TOO_LONG",
			Value:   username,
		}
	}

	if !usernameRegex.MatchString(username) {
		return &ValidationError{
			Field:   "username",
			Message: "Username can only contain letters, numbers, underscores and hyphens",
			Code:    "INVALID_FORMAT",
			Value:   username,
		}
	}

	// Проверяем, что username не состоит только из цифр
	if regexp.MustCompile(`^\d+$`).MatchString(username) {
		return &ValidationError{
			Field:   "username",
			Message: "Username cannot consist only of numbers",
			Code:    "INVALID_FORMAT",
			Value:   username,
		}
	}

	return nil
}

// Password validation
func ValidatePassword(password string) *ValidationError {
	if password == "" {
		return &ValidationError{
			Field:   "password",
			Message: "Password is required",
			Code:    "REQUIRED",
		}
	}

	if len(password) < 6 {
		return &ValidationError{
			Field:   "password",
			Message: "Password must be at least 6 characters long",
			Code:    "TOO_SHORT",
		}
	}

	if len(password) > 128 {
		return &ValidationError{
			Field:   "password",
			Message: "Password is too long (maximum 128 characters)",
			Code:    "TOO_LONG",
		}
	}

	return nil
}

// Strong password validation
func ValidateStrongPassword(password string) *ValidationError {
	if err := ValidatePassword(password); err != nil {
		return err
	}

	if len(password) < 8 {
		return &ValidationError{
			Field:   "password",
			Message: "Password must be at least 8 characters long",
			Code:    "TOO_SHORT",
		}
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return &ValidationError{
			Field:   "password",
			Message: "Password must contain at least one uppercase letter",
			Code:    "MISSING_UPPERCASE",
		}
	}

	if !hasLower {
		return &ValidationError{
			Field:   "password",
			Message: "Password must contain at least one lowercase letter",
			Code:    "MISSING_LOWERCASE",
		}
	}

	if !hasDigit {
		return &ValidationError{
			Field:   "password",
			Message: "Password must contain at least one digit",
			Code:    "MISSING_DIGIT",
		}
	}

	if !hasSpecial {
		return &ValidationError{
			Field:   "password",
			Message: "Password must contain at least one special character",
			Code:    "MISSING_SPECIAL",
		}
	}

	return nil
}

// Room name validation
func ValidateRoomName(name string) *ValidationError {
	if name == "" {
		return &ValidationError{
			Field:   "name",
			Message: "Room name is required",
			Code:    "REQUIRED",
		}
	}

	// Убираем лишние пробелы
	name = strings.TrimSpace(name)

	if len(name) < 3 {
		return &ValidationError{
			Field:   "name",
			Message: "Room name must be at least 3 characters long",
			Code:    "TOO_SHORT",
			Value:   name,
		}
	}

	if len(name) > 255 {
		return &ValidationError{
			Field:   "name",
			Message: "Room name must be at most 255 characters long",
			Code:    "TOO_LONG",
			Value:   name,
		}
	}

	if !roomNameRegex.MatchString(name) {
		return &ValidationError{
			Field:   "name",
			Message: "Room name can only contain letters, numbers, spaces, hyphens and underscores",
			Code:    "INVALID_FORMAT",
			Value:   name,
		}
	}

	return nil
}

// Max players validation
func ValidateMaxPlayers(maxPlayers int) *ValidationError {
	if maxPlayers < 2 {
		return &ValidationError{
			Field:   "max_players",
			Message: "Maximum players must be at least 2",
			Code:    "TOO_SMALL",
			Value:   maxPlayers,
		}
	}

	if maxPlayers > 64 {
		return &ValidationError{
			Field:   "max_players",
			Message: "Maximum players cannot exceed 64",
			Code:    "TOO_LARGE",
			Value:   maxPlayers,
		}
	}

	return nil
}

// Hero name validation
func ValidateHeroName(name string) *ValidationError {
	if name == "" {
		return &ValidationError{
			Field:   "name",
			Message: "Hero name is required",
			Code:    "REQUIRED",
		}
	}

	name = strings.TrimSpace(name)

	if len(name) < 2 {
		return &ValidationError{
			Field:   "name",
			Message: "Hero name must be at least 2 characters long",
			Code:    "TOO_SHORT",
			Value:   name,
		}
	}

	if len(name) > 100 {
		return &ValidationError{
			Field:   "name",
			Message: "Hero name must be at most 100 characters long",
			Code:    "TOO_LONG",
			Value:   name,
		}
	}

	return nil
}

// Hero element validation
func ValidateHeroElement(element string) *ValidationError {
	validElements := []string{"Physical", "Fire", "Ice", "Electric", "Ether"}

	if element == "" {
		return &ValidationError{
			Field:   "element",
			Message: "Hero element is required",
			Code:    "REQUIRED",
		}
	}

	for _, valid := range validElements {
		if element == valid {
			return nil
		}
	}

	return &ValidationError{
		Field:   "element",
		Message: "Invalid hero element. Must be one of: " + strings.Join(validElements, ", "),
		Code:    "INVALID_VALUE",
		Value:   element,
	}
}

// Hero rarity validation
func ValidateHeroRarity(rarity string) *ValidationError {
	validRarities := []string{"A", "S"}

	if rarity == "" {
		return &ValidationError{
			Field:   "rarity",
			Message: "Hero rarity is required",
			Code:    "REQUIRED",
		}
	}

	for _, valid := range validRarities {
		if rarity == valid {
			return nil
		}
	}

	return &ValidationError{
		Field:   "rarity",
		Message: "Invalid hero rarity. Must be one of: " + strings.Join(validRarities, ", "),
		Code:    "INVALID_VALUE",
		Value:   rarity,
	}
}

// Hero role validation
func ValidateHeroRole(role string) *ValidationError {
	validRoles := []string{"Attack", "Stun", "Anomaly", "Support", "Defense"}

	if role == "" {
		return &ValidationError{
			Field:   "role",
			Message: "Hero role is required",
			Code:    "REQUIRED",
		}
	}

	for _, valid := range validRoles {
		if role == valid {
			return nil
		}
	}

	return &ValidationError{
		Field:   "role",
		Message: "Invalid hero role. Must be one of: " + strings.Join(validRoles, ", "),
		Code:    "INVALID_VALUE",
		Value:   role,
	}
}

// URL validation
func ValidateURL(url string) *ValidationError {
	if url == "" {
		return nil // URL не обязательный
	}

	if len(url) > 2000 {
		return &ValidationError{
			Field:   "url",
			Message: "URL is too long (maximum 2000 characters)",
			Code:    "TOO_LONG",
			Value:   url,
		}
	}

	if !urlRegex.MatchString(url) {
		return &ValidationError{
			Field:   "url",
			Message: "Invalid URL format",
			Code:    "INVALID_FORMAT",
			Value:   url,
		}
	}

	return nil
}

// Phone number validation
func ValidatePhoneNumber(phone string) *ValidationError {
	if phone == "" {
		return nil // Phone не обязательный
	}

	// Убираем пробелы и тире
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	if !phoneRegex.MatchString(phone) {
		return &ValidationError{
			Field:   "phone",
			Message: "Invalid phone number format",
			Code:    "INVALID_FORMAT",
			Value:   phone,
		}
	}

	return nil
}

// Rating validation
func ValidateRating(rating int) *ValidationError {
	if rating < 0 {
		return &ValidationError{
			Field:   "rating",
			Message: "Rating cannot be negative",
			Code:    "TOO_SMALL",
			Value:   rating,
		}
	}

	if rating > 4000 {
		return &ValidationError{
			Field:   "rating",
			Message: "Rating cannot exceed 4000",
			Code:    "TOO_LARGE",
			Value:   rating,
		}
	}

	return nil
}

// Text length validation
func ValidateTextLength(text, fieldName string, minLength, maxLength int) *ValidationError {
	if text == "" && minLength > 0 {
		return &ValidationError{
			Field:   fieldName,
			Message: fieldName + " is required",
			Code:    "REQUIRED",
		}
	}

	text = strings.TrimSpace(text)

	if len(text) < minLength {
		return &ValidationError{
			Field:   fieldName,
			Message: fieldName + " must be at least " + fmt.Sprintf("%d", minLength) + " characters long",
			Code:    "TOO_SHORT",
			Value:   text,
		}
	}

	if len(text) > maxLength {
		return &ValidationError{
			Field:   fieldName,
			Message: fieldName + " must be at most " + fmt.Sprintf("%d", maxLength) + " characters long",
			Code:    "TOO_LONG",
			Value:   text,
		}
	}

	return nil
}

// ID validation
func ValidateID(id int, fieldName string) *ValidationError {
	if id <= 0 {
		return &ValidationError{
			Field:   fieldName,
			Message: fieldName + " must be a positive integer",
			Code:    "INVALID_VALUE",
			Value:   id,
		}
	}

	return nil
}

// Page validation for pagination
func ValidatePage(page int) *ValidationError {
	if page < 1 {
		return &ValidationError{
			Field:   "page",
			Message: "Page must be at least 1",
			Code:    "TOO_SMALL",
			Value:   page,
		}
	}

	return nil
}

// PerPage validation for pagination
func ValidatePerPage(perPage int) *ValidationError {
	if perPage < 1 {
		return &ValidationError{
			Field:   "per_page",
			Message: "Per page must be at least 1",
			Code:    "TOO_SMALL",
			Value:   perPage,
		}
	}

	if perPage > 100 {
		return &ValidationError{
			Field:   "per_page",
			Message: "Per page cannot exceed 100",
			Code:    "TOO_LARGE",
			Value:   perPage,
		}
	}

	return nil
}

// Batch validation helper
func ValidateStruct(validations map[string]func() *ValidationError) ValidationErrors {
	var errors ValidationErrors

	for _, validate := range validations {
		if err := validate(); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// Common validation sets
func ValidateUserRegistration(username, email, password string) ValidationErrors {
	return ValidateStruct(map[string]func() *ValidationError{
		"username": func() *ValidationError { return ValidateUsername(username) },
		"email":    func() *ValidationError { return ValidateEmail(email) },
		"password": func() *ValidationError { return ValidatePassword(password) },
	})
}

func ValidateUserLogin(username, password string) ValidationErrors {
	return ValidateStruct(map[string]func() *ValidationError{
		"username": func() *ValidationError { return ValidateUsername(username) },
		"password": func() *ValidationError { return ValidatePassword(password) },
	})
}

func ValidateRoomCreation(name string, maxPlayers int) ValidationErrors {
	return ValidateStruct(map[string]func() *ValidationError{
		"name":        func() *ValidationError { return ValidateRoomName(name) },
		"max_players": func() *ValidationError { return ValidateMaxPlayers(maxPlayers) },
	})
}

func ValidateHeroCreation(name, element, rarity, role string) ValidationErrors {
	return ValidateStruct(map[string]func() *ValidationError{
		"name":    func() *ValidationError { return ValidateHeroName(name) },
		"element": func() *ValidationError { return ValidateHeroElement(element) },
		"rarity":  func() *ValidationError { return ValidateHeroRarity(rarity) },
		"role":    func() *ValidationError { return ValidateHeroRole(role) },
	})
}
