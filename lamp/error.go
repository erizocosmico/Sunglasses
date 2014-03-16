package lamp

const (
	// Misc codes
	CodeUnexpected  = 10
	CodeInvalidData = 11

	// Auth codes
	CodeInvalidAccessToken        = 1
	CodeInvalidUserToken          = 2
	CodeTokenNotFound             = 3
	CodeInvalidUsernameOrPassword = 4

	// User codes
	CodeUserDoesNotExist          = 20
	CodeUserCantBeRequested       = 21
	CodeFollowRequestDoesNotExist = 22

	// Misc messages
	MsgUnexpected  = "Unexpected error occurred"
	MsgInvalidData = "Invalid data provided"

	// Auth messages
	MsgInvalidAccessToken        = "Invalid access token provided"
	MsgInvalidUserToken          = "Invalid user token provided"
	MsgTokenNotFound             = "Token not found"
	MsgInvalidUsernameOrPassword = "Invalid username or password"

	// User messages
	MsgUserDoesNotExist          = "Requested user does not exist"
	MsgUserCantBeRequested       = "You can't send a follow request to that user"
	MsgFollowRequestDoesNotExist = "The follow request does not exist"
)
