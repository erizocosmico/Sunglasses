package mask

const (
	// Misc codes [10-19]
	CodeUnexpected       = 10
	CodeInvalidData      = 11
	CodeUnauthorized     = 12
	CodeNotFound         = 13
	CodeInvalidSignature = 14

	// Auth codes [1-9]
	CodeInvalidAccessToken        = 1
	CodeInvalidUserToken          = 2
	CodeTokenNotFound             = 3
	CodeInvalidUsernameOrPassword = 4

	// User codes [20-49]
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
	CodeInvalidGender             = 30
	CodeInvalidStatus             = 31
	CodeInvalidWebsites           = 32
	CodeInvalidInfoLength         = 33
	CodeInvalidPrivacySettings    = 34

	// Post Codes [50-70]
	CodeInvalidStatusText     = 50
	CodeInvalidCaption        = 51
	CodeFileTooLarge          = 52
	CodeNoFileUploaded        = 53
	CodeInvalidFileFormat     = 54
	CodeInvalidFileDimensions = 55
	CodeInvalidFile           = 56

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
	MsgInvalidGender             = "Invalid gender"
	MsgInvalidStatus             = "Invalid user status"
	MsgInvalidWebsites           = "One or more of the provided websites is not a valid url"
	MsgInvalidInfoLength         = "One or more of the provided fields is more than 500 characters long"
	MsgInvalidPrivacySettings    = "Invalid privacy settings provided"

	// Post messages
	MsgInvalidStatusText     = "Status text must not be more than 1500 characters long"
	MsgInvalidCaption        = "Caption must not be more than 255 characters long"
	MsgFileTooLarge          = "The maximum file size allowed is 10MB"
	MsgNoFileUploaded        = "No file was uploaded"
	MsgInvalidFileFormat     = "Invalid file format"
	MsgInvalidFileDimensions = "Invalid file dimensiones"
	MsgInvalidFile           = "Invalid file"
)
