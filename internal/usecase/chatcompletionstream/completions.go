package chatcompletionstream

import (
	"context"
	"errors"
	"io"
	"strings"

	"git.com/msviniicius/chatGPTservice/internal/domain/entity"
	"git.com/msviniicius/chatGPTservice/internal/domain/gateway"
	openai "github.com/sashabaranov/go-openai"
)

// objeto que guarda valores e não tem um coportamento
// são dados primitivos
// dados para configuração
type ChatCompletionConfigImputDTO struct {
	Model                string
	ModelMaxTokenx       int
	Temperature          float32
	TopP                 float32
	N                    int
	Stop                 []string
	MaxTokens            int
	PresencePenalty      float32
	FrequencyPenaty      float32
	InicialSystemMessage string
}

// dados de entrada
// dados de complicion que o user vai mandar
type ChatCompletionImputDTO struct {
	ChatID      string
	UserID      string
	UserMessage string
	Config      ChatCompletionConfigImputDTO
}

// dados de saida
type ChatCompletionOutputDTO struct {
	ChatID  string
	UserID  string
	Content string
}

type ChatCompletionUseCase struct {
	ChatGateway  gateway.ChatGateway          // versão de controle, onde terá o inject dependia do ChatGateway | salvar o dado no Banco de Dados
	OpenAiClient *openai.Client               // acessa o openai | chamar a Api
	Stream       chan ChatCompletionOutputDTO // canal de comunicação, conforme vai recebendo, vai pegando os dados e enviando para outra thedres
}

func NewChatCompletionUseCase(chatGateway gateway.ChatGateway, openAiClient *openai.Client, stream chan ChatCompletionOutputDTO) *ChatCompletionUseCase {
	return &ChatCompletionUseCase{
		ChatGateway:  chatGateway,
		OpenAiClient: openAiClient,
	}
}

func (uc *ChatCompletionUseCase) Execute(ctx context.Context, input ChatCompletionImputDTO) (*ChatCompletionOutputDTO, error) {
	chat, err := uc.ChatGateway.FindChatID(ctx, input.ChatID)
	if err != nil {
		if err.Error() == "Chat Not found" {
			chat, err = createNewChat(input)
			if err != nil {
				return nil, errors.New("error creating new chat: " + err.Error())
			}

			// save database
			err = uc.ChatGateway.CreateChat(ctx, chat)
			if err != nil {
				return nil, errors.New("error persisting new chat: " + err.Error())
			}
		} else {
			return nil, errors.New("error fetching existing chat: " + err.Error())
		}
	}

	userMessage, err := entity.NewMessage("user", input.UserMessage, chat.Config.Model)
	if err != nil {
		return nil, errors.New("error creating user message: " + err.Error())
	}

	err = chat.AddMessage(userMessage)
	if err != nil {
		return nil, errors.New("error adding new message: " + err.Error())
	}

	// recebendo todas as messages do chat
	messages := []openai.ChatCompletionMessage{}
	for _, msg := range chat.Messages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	resp, err := uc.OpenAiClient.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:            chat.Config.Model.Name,
			Messages:         messages,
			MaxTokens:        chat.Config.MaxTokens,
			Temperature:      chat.Config.Temperature,
			TopP:             chat.Config.TopP,
			PresencePenalty:  chat.Config.PresentPenalty,
			FrequencyPenalty: chat.Config.FrequencyPenalty,
			Stop:             chat.Config.Stop,
			Stream:           true,
		},
	)

	if err != nil {
		return nil, errors.New("error creating chat completion: " + err.Error())
	}

	var fullResponse strings.Builder

	for {
		// recebo os dados via streaming
		response, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, errors.New("error streaming response: " + err.Error())
		}

		fullResponse.WriteString(response.Choices[0].Delta.Content)
		r := ChatCompletionOutputDTO{
			ChatID:  chat.ID,
			UserID:  input.UserID,
			Content: fullResponse.String(),
		}
		// mandando para o canal de streaming
		uc.Stream <- r
	}

	// responsavel por manter o historico
	assistent, err := entity.NewMessage("assistent", fullResponse.String(), chat.Config.Model)
	if err != nil {
		return nil, errors.New("error creating assistent message: " + err.Error())
	}

	err = chat.AddMessage(assistent)
	if err != nil {
		return nil, errors.New("error adding assistent message: " + err.Error())
	}
	// save o chat
	err = uc.ChatGateway.SaveChat(ctx, chat)
	if err != nil {
		return nil, errors.New("error saving chat: " + err.Error())
	}

	return &ChatCompletionOutputDTO{
		ChatID:  chat.ID,
		UserID:  input.UserID,
		Content: fullResponse.String(),
	}, nil
}

func createNewChat(input ChatCompletionImputDTO) (*entity.Chat, error) {
	model := entity.NewModel(input.Config.Model, input.Config.ModelMaxTokenx)
	chatConfig := &entity.ChatConfig{
		Temperature:      input.Config.Temperature,
		TopP:             input.Config.TopP,
		N:                input.Config.N,
		Stop:             input.Config.Stop,
		MaxTokens:        input.Config.MaxTokens,
		PresentPenalty:   input.Config.PresencePenalty,
		FrequencyPenalty: input.Config.FrequencyPenaty,
		Model:            model,
	}

	initialMessage, err := entity.NewMessage("system", input.Config.InicialSystemMessage, model)
	if err != nil {
		return nil, errors.New("error creating initial message: " + err.Error())
	}

	chat, err := entity.NewChat(input.UserID, initialMessage, chatConfig)
	if err != nil {
		return nil, errors.New("error creating new chat" + err.Error())
	}

	return chat, nil
}
