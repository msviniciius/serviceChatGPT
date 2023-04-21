package entity

type ChatConfig struct {
	Model            *Model
	Temperature      float32  // quanto mais perto de 0.0 mais preciso é a resposta
	TopP             float32  // quanto mais conservador ele será na escolha das mensagens (palavras)
	N                int      // numero de msg maxima que ele pode gerar
	Stop             []string // quando parar o chat
	MaxTokens        int      // quantos tokens essa conversa poderá ter
	PresentPenalty   float32
	FrequencyPenalty float32
}

type Chat struct {
	ID                   string
	UserID               string
	InitialSystemMessage *Message
	Message              []*Message
	ArasedMessage        []*Message
	Status               string
	TokenUsage           int
	Config               *ChatConfig
}
