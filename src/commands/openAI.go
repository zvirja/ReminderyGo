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
		return cmd.createReminderFromText(ctx, ctx.Text())
	})

	bot.Handle(tele.OnVoice, func(ctx tele.Context) error {
		return cmd.createReminderFromVoiceMsg(ctx)
	})
}

func (cmd OpenAICmd) createReminderFromVoiceMsg(ctx tele.Context) error {
	voice := ctx.Message().Voice

	file, err := ctx.Bot().File(&voice.File)
	if err != nil {
		return err
	}

	defer file.Close()

	client := cmd.createOpenAiClient()

	resp, err := client.CreateTranscription(context.Background(), openai.AudioRequest{
		Model:  openai.Whisper1,
		Reader: file,
		// Format:   openai.AudioResponseFormatText,
		FilePath: voice.FilePath,
		Language: "en",
	})

	if err != nil {
		err = ctx.Send(fmt.Sprintf("Failed to get OpenAI transcription reply: %v", err.Error()))
		if err != nil {
			return err
		}

		return nil
	}

	text := resp.Text
	slog.Debug("Received OpenAI transcription response", "text", text)

	return cmd.createReminderFromText(ctx, text)
}

func (cmd OpenAICmd) createReminderFromText(ctx tele.Context, text string) error {
	client := cmd.createOpenAiClient()

	request := openai.ChatCompletionRequest{
		// Model: "gpt-3.5-turbo-1106",
		Model: "gpt-3.5-turbo",
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
				Content: text,
			},
		},
		Tools: []openai.Tool{
			{
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionDefinition{
					Name:        "create_reminder",
					Description: "Creates reminder with given message",
					Parameters: json.RawMessage(`
						{
							"type": "object",
							"properties": {
								"Message": {
									"type": "string",
									"description": "Message to remind about"
								},
								"Schedule": {
									"type": "array",
									"description": "Specifies one or more schedules when to show current reminder. If more than one is specified, reminder is shown on any of them.",
									"items": {
										"type": "object",
										"properties": {
											"Kind": {
												"type": "string",
												"description": "Specifies kind of schedule",
												"enum": [
													"one-time",
													"recurrent"
												]
											},
											"Cron": {
												"type": "string",
												"description": "Valid Cron schedule definition to describe when reminder should be shown. Should only be used if \"Kind\" property is \"recurrent\". Otherwise should be omitted"
											},
											"DateTime": {
												"type": "string",
												"description": "Date and time in ISO 8601. Assume time zone is always UTC. Should only be used if \"Kind\" property is \"one-time\". Otherwise should be omitted"
											}
										},
										"required": [
											"Kind"
										]
									}
								}
							},
							"required": [
								"Message",
								"Schedule"
							]
						}
					`),
				},
			},
		},
	}
	slog.Debug("Make OpenAI request", "request", text)

	response, err := client.CreateChatCompletion(context.Background(), request)
	slog.Debug("Received OpenAI response", "choices", response.Choices)

	if err != nil {
		err = ctx.Send(fmt.Sprintf("Failed to get OpenAI completion reply: %v", err))
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
			case "create_reminder":
				if err = createRemider(ctx, toolCall.Function); err != nil {
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
}

func createRemider(ctx tele.Context, call openai.FunctionCall) error {
	return ctx.Send(fmt.Sprintf("Creating reminder: %v", call.Arguments))
}
