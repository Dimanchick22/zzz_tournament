// internal/models/constants.go
package models

// Database constraints
const (
	// User constraints
	MinUsernameLength = 3
	MaxUsernameLength = 50
	MinPasswordLength = 8
	MaxPasswordLength = 128

	// Room constraints
	MinRoomNameLength     = 3
	MaxRoomNameLength     = 100
	MaxRoomDescription    = 500
	MinPlayersInRoom      = 2
	MaxPlayersInRoom      = 32
	MaxRoomPasswordLength = 50

	// Message constraints
	MaxMessageLength = 1000

	// Tournament constraints
	MinTournamentNameLength = 3
	MaxTournamentNameLength = 100

	// Hero constraints
	MaxHeroNameLength  = 50
	MaxHeroDescription = 1000
)

// Default values
const (
	DefaultUserRating = 1000
	DefaultMaxPlayers = 8
)

// Error codes
const (
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeUserNotFound       = "USER_NOT_FOUND"
	ErrCodeUserAlreadyExists  = "USER_ALREADY_EXISTS"
	ErrCodeRoomNotFound       = "ROOM_NOT_FOUND"
	ErrCodeRoomFull           = "ROOM_FULL"
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeForbidden          = "FORBIDDEN"
	ErrCodeValidationFailed   = "VALIDATION_FAILED"
	ErrCodeInternalError      = "INTERNAL_ERROR"
)
