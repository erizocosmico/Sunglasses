package mask

const (
	// Misc codes
	CodeUnexpected       = 10
	CodeInvalidData      = 11
	CodeUnauthorized     = 12
	CodeNotFound         = 13
	CodeInvalidSignature = 14

	// Auth codes
	CodeInvalidAccessToken        = 1
	CodeInvalidUserToken          = 2
	CodeTokenNotFound             = 3
	CodeInvalidUsernameOrPassword = 4

	// User codes
	CodeUserDoesNotExist          = 20
	CodeUserCantBeRequested       = 21
	CodeFollowRequestDoesNotExist = 22
	CodeUsernameTaken             = 23
	CodeInvalidEmail              = 24
	CodeInvalidRecoveryQuestion   = 25
	CodeInvalidRecoveryMethod     = 26
	CodeInvalidUsername           = 27
	CodePasswordLength            = 28
	CodePasswordMatch             = 29

	// Misc messages
	MsgUnexpected       = "Unexpected error occurred"
	MsgInvalidData      = "Invalid data provided"
	MsgUnauthorized     = "You are not authorized to access this resource"
	MsgNotFound         = "The resource was not found"
	MsgInvalidSignature = "Invalid signature"

	// Auth messages
	MsgInvalidAccessToken        = "Invalid access token provided"
	MsgInvalidUserToken          = "Invalid user token provided"
	MsgTokenNotFound             = "Token not found"
	MsgInvalidUsernameOrPassword = "Invalid username or password"

	// User messages
	MsgUserDoesNotExist          = "Requested user does not exist"
	MsgUserCantBeRequested       = "You can't send a follow request to that user"
	MsgFollowRequestDoesNotExist = "The follow request does not exist"
	MsgUsernameTaken             = "The username is taken"
	MsgInvalidEmail              = "Invalid email address"
	MsgInvalidRecoveryQuestion   = "Neither the recovery question nor the recovery answer can be blank"
	MsgInvalidRecoveryMethod     = "Invalid recovery method"
	MsgInvalidUsername           = "Invalid username"
	MsgPasswordLength            = "Password must be at least 6 characters long"
	MsgPasswordMatch             = "Passwords don't match"
)
