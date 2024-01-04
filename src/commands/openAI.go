package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	tele "gopkg.in/telebot.v3"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAICmd struct {
	Token string
}

func (cmd OpenAICmd) createOpenAiClient() *openai.Client {
	return openai.NewClient(cmd.Token)
}

func (cmd OpenAICmd) RegisterHandler(bot *tele.Bot) {
	bot.Handle(tele.OnText, func(ctx tele.Context) error {
		client := cmd.createOpenAiClient()

		request := openai.ChatCompletionRequest{
			Model: "gpt-3.5-turbo-1106",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "Try to invoke function to create a reminder. If you invoke a function - reply nothing else. If cannot - reply with a really short message to clarify missing parts to build a reminder",
				},
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "If something has to be done in the morning and time is not specified, pick random time between 10:00AM and 11:00AM",
				},
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: fmt.Sprintf("Date and time now: %v", time.Now().Local().Format(time.RFC3339)),
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: ctx.Text(),
				},
			},
			Tools: []openai.Tool{
				{
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionDefinition{
						Name:        "create_one_time_reminder",
						Description: "Creates a one time reminder in future with given message",
						Parameters: json.RawMessage(`
							{
								"type": "object",
								"properties": {
									"DateTime": {
										"type": "string",
										"description": "Date and time in ISO 8601. Assume time zone is always UTC"
									},
									"Message": {
										"type": "string",
										"description": "Message to remind about"
									}
								},
								"required": [
									"DateTime",
									"Message"
								]
							}
						`),
					},
				},
				{
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionDefinition{
						Name:        "create_recurrent_reminder",
						Description: "Creates a recurrent reminder with given message",
						Parameters: json.RawMessage(`
							{
								"type": "object",
								"properties": {
									"Schedules": {
										"type": "array",
										"description": "List of valid schedules to remind at. Notice, could contain multiple schedules - will be notified when any of them fires.",
										"items": {
											"type": "object",
											"properties": {
												"Cron": {
													"type": "string",
													"description": "Valid Cron schedule definition to describe when reminder should be fired"
												}
											},
											"required": [
												"Cron"
											]
										}
									},
									"Message": {
										"type": "string",
										"description": "Message to remind about"
									}
								},
								"required": [
									"Schedules",
									"Message"
								]
							}
						`),
					},
				},
			},
		}
		slog.Debug("Make OpenAI request", "request", ctx.Text())

		response, err := client.CreateChatCompletion(context.Background(), request)
		slog.Debug("Received OpenAI response", "choices", response.Choices)

		if err != nil {
			err = ctx.Send(fmt.Sprintf("Failed to get OpenAI reply: %v", err.Error()))
			if err != nil {
				return err
			}

			return nil
		}

		for _, choice := range response.Choices {
			hadToolCall := false

			for _, toolCall := range choice.Message.ToolCalls {
				hadToolCall = true

				switch toolCall.Function.Name {
				case "create_one_time_reminder":
					if err = createOneTimeRemider(ctx, toolCall.Function); err != nil {
						return err
					}

				case "create_recurrent_reminder":
					if err = createRecurrentRemider(ctx, toolCall.Function); err != nil {
						return err
					}
				}
			}

			if !hadToolCall {
				if err = ctx.Send(fmt.Sprintf("🗿 OpenAI reply: %v", choice.Message.Content)); err != nil {
					return err
				}

			}
		}

		return nil
	})

	bot.Handle(tele.OnVoice, func(ctx tele.Context) error {
		voice := ctx.Message().Voice

		file, err := bot.File(&voice.File)
		if err != nil {
			return err
		}

		defer file.Close()

		client := cmd.createOpenAiClient()

		resp, err := client.CreateTranscription(context.Background(), openai.AudioRequest{
			Model:    openai.Whisper1,
			Reader:   file,
			Format:   openai.AudioResponseFormatText,
			FilePath: voice.FilePath,
			// Language: "en",
		})

		if err != nil {
			err = ctx.Send(fmt.Sprintf("Failed to get OpenAI reply: %v", err.Error()))
			if err != nil {
				return err
			}

			return nil
		}

		slog.Debug("Received OpenAI response", "response", resp)

		return ctx.Send(fmt.Sprintf("Text: %v", resp.Text))
	})
}

func createOneTimeRemider(ctx tele.Context, call openai.FunctionCall) error {
	return ctx.Send(fmt.Sprintf("Creating one time reminder: %v", call.Arguments))
}

func createRecurrentRemider(ctx tele.Context, call openai.FunctionCall) error {
	return ctx.Send(fmt.Sprintf("Creating recurrent reminder: %v", call.Arguments))
}
