package audiohook_genesys

import "encoding/json"

// MessageReceived représente la structure des messages reçus
type MessageReceived struct {
	Version    string          `json:"version"`
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Seq        int             `json:"seq"`
	ServerSeq  int             `json:"serverseq"`
	Position   string          `json:"position"` // Vous pouvez utiliser un type personnalisé pour Duration
	Parameters json.RawMessage `json:"parameters"`
}

// MessageSent représente la structure des messages envoyés
type MessageSent struct {
	Version    string      `json:"version"`
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Seq        int         `json:"seq"`
	ClientSeq  int         `json:"clientseq"`
	Parameters interface{} `json:"parameters"`
}

// OpenParameters représente les paramètres pour le message "open"
type OpenParameters struct {
	OrganizationId string                 `json:"organizationId"`
	ConversationId string                 `json:"conversationId"`
	Participant    Participant            `json:"participant"`
	Language       *string                `json:"language,omitempty"`
	CustomConfig   map[string]interface{} `json:"customConfig,omitempty"`
	InputVariables map[string]string      `json:"inputVariables,omitempty"`
}

// Participant représente les informations du participant
type Participant struct {
	ID      string `json:"id"`
	Ani     string `json:"ani"`
	AniName string `json:"aniName"`
	Dnis    string `json:"dnis"`
}

// OpenedParameters représente les paramètres pour le message "opened"
type OpenedParameters struct {
	Media       []MediaParameter `json:"media"`
	DiscardTo   *string          `json:"discardTo,omitempty"`
	StartPaused bool             `json:"startPaused,omitempty"`
}

// MediaParameter représente un paramètre de média
type MediaParameter struct {
	Type     string   `json:"type"`     // "audio"
	Format   string   `json:"format"`   // "PCMU"
	Channels []string `json:"channels"` // ["external", "internal"]
	Rate     int      `json:"rate"`     // 8000
}

// UpdateParameters représente les paramètres pour le message "update"
type UpdateParameters struct {
	Language *string `json:"language,omitempty"`
}

// PingParameters représente les paramètres pour le message "ping"
type PingParameters struct {
	RTT *string `json:"rtt,omitempty"` // Vous pouvez utiliser un type personnalisé pour Duration
}

// PongParameters représente les paramètres pour le message "pong"
// Ici, aucun paramètre n'est nécessaire, mais la structure est maintenue pour la cohérence
type PongParameters struct{}

// ClosedParameters représente les paramètres pour le message "closed"
type ClosedParameters struct{}
