package entity

import "errors"

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
	Messages             []*Message
	ArasedMessages       []*Message
	Status               string
	TokenUsage           int
	Config               *ChatConfig
}

func NewChat(userID string, initialSystemMessage *Message, ChatConfig *ChatConfig) (*Chat, error) {

}

func (c *Chat) Validate() error {
	if c.UserID == "" {
		return errors.New("user ID is empty")
	}

	if c.Status != "active" && c.Status != "ended" {
		return errors.New("invalid status")
	}

	if c.Config.Temperature < 0 || c.Config.Temperature > 2 {
		return errors.New("invalid temperature")
	}

	return nil
}

func (c *Chat) AddMessage(m *Message) error {
	if c.Status == "ended" {
		return errors.New("chat is already ended")
	}

	// verifico a quantidade de tokens que o model suporta
	// verifico a quantidade de tokens da mensagem
	// verifico a quantidade de uso de token do chat
	for {
		if c.Config.Model.GetMaxTokens() >= m.GetQtdTokens()+c.TokenUsage {
			c.Messages = append(c.Messages, m)
			c.RefreshTokenUsade()
			break
		}
		// caso o espaço do nosso modelo estiver cheio
		// pego a ultima mensagem e apago
		c.ArasedMessages = append(c.ArasedMessages, c.Messages[0])
		c.Messages = c.Messages[1:]
		c.RefreshTokenUsade()
	}

	return nil
}

func (c *Chat) GetMessages() []*Message {
	return c.Messages
}

func (c *Chat) CountMessages() int {
	return len(c.Messages)
}

func (c *Chat) End() {
	c.Status = "ended"
}

func (c *Chat) RefreshTokenUsade() {
	c.TokenUsage = 0

	for m := range c.Messages {
		c.TokenUsage += c.Messages[m].GetQtdTokens()
	}
}
