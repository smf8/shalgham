package command

const (
	TypeLogin              = "login"
	TypeSignup             = "signup"
	TypeSendText           = "send-text"
	TypeUserStatus         = "user-status"
	TypeConversationStatus = "conversation-status"
	TypeJoinConversation   = "join-conversation"
	TypeChangeUsername     = "change-username"
	TypeFileMessage        = "file-message"

	FileChunkSize = 50 * 1024
)
