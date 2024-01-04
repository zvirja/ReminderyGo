package commands

import (
	"log/slog"
	"reflect"
	"remindery/environment"

	tele "gopkg.in/telebot.v3"
)

type commandDescriptor struct {
	Text        string
	Description string
}

type commandMenuer interface {
	Descriptor() commandDescriptor
}

type commandHandler interface {
	RegisterHandler(bot *tele.Bot)
}

func SetupCommands(bot *tele.Bot, env *environment.Environment) (err error) {
	commands := [...]interface{}{
		StartCmd{
			Version: env.Version,
		},
		OpenAICmd{
			Token: env.Config.OpenAIToken,
		},
	}

	// Register commands in Menu
	{
		botCommands := make([]tele.Command, 0, 10)
		for _, cmd := range commands {
			switch value := cmd.(type) {
			case commandMenuer:
				desc := value.Descriptor()
				botCommands = append(botCommands, tele.Command{
					Text:        desc.Text,
					Description: desc.Description,
				})
			}
		}

		err = bot.SetCommands(botCommands)
		if err != nil {
			return
		}

		slog.Debug("Registered commands in Menu", "count", len(botCommands))
	}

	// Register commands handlers
	{
		for _, cmd := range commands {
			switch value := cmd.(type) {
			case commandHandler:
				value.RegisterHandler(bot)
				slog.Debug("Registered handler for command", "command", reflect.TypeOf(value).Name())
			}
		}
	}

	return
}
