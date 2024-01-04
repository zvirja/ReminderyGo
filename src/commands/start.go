package commands

import (
	"fmt"
	"remindery/environment"

	tele "gopkg.in/telebot.v3"
)

type StartCmd struct {
	Version environment.AppVersion
}

func (cmd StartCmd) RegisterHandler(bot *tele.Bot) {
	bot.Handle("/start", func(ctx tele.Context) error {
		return ctx.Send(fmt.Sprintf("This is Remindery bot v%v (%v)!", cmd.Version.Version, cmd.Version.GitCommit))
	})
}
