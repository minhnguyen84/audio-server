package audiohook_genesys

const (
	// Types de messages reçus
	MessageTypeOpen   = "open"
	MessageTypeUpdate = "update"
	MessageTypePing   = "ping"
	MessageTypeClose  = "close"

	// Types de messages envoyés
	MessageTypeOpened = "opened"
	MessageTypePong   = "pong"
	MessageTypeClosed = "closed"
	MessageTypeError  = "error"

	// Paramètres spécifiques aux messages
	MediaTypeAudio       = "audio"
	MediaFormatPCMU      = "PCMU"
	MediaChannelExternal = "external"
	MediaChannelInternal = "internal"
	MediaRate            = 8000
)
