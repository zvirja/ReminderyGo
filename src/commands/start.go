package commands

import (
	"fmt"

	tele "gopkg.in/telebot.v3"
)

type StartCmd struct {
	Version string
	Commit  string
}

func (cmd StartCmd) RegisterHandler(bot *tele.Bot) {
	bot.Handle("/start", func(ctx tele.Context) error {
		return ctx.Send(fmt.Sprintf("This is Remindery bot v%v (%v)!", cmd.Version, cmd.Commit))
	})
}
