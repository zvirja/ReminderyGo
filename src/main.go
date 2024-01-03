package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"remindery/commands"
	"remindery/environment"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

var (
	version   string = "0.1.0"
	commitSHA string = "aabbcc"
)

func main() {
	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, os.Interrupt)

	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}

	botToken, ok := os.LookupEnv("BOT_TOKEN")
	if !ok {
		log.Fatal("BOT_TOKEN env variable is not set")
	}

	pref := tele.Settings{
		Token:       botToken,
		Poller:      &tele.LongPoller{},
		Synchronous: true,
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	bot.Use(middleware.Logger(slog.NewLogLogger(slog.Default().Handler(), slog.LevelDebug)))

	env := environment.Environment{
		Version: version,
		Commit:  commitSHA,
	}

	err = commands.SetupCommands(bot, &env)
	if err != nil {
		slog.Error("Unable to register commads", "error", err)
		return
	}

	bot.Handle("/hello", func(ctx tele.Context) error {
		return ctx.Send("HELLO")
	})

	slog.Info("Started bot", "version", version)
	go func() {
		<-interruptSignal
		slog.Debug("Received INT signal. Stopping...")
		bot.Stop()
	}()

	bot.Start()
	slog.Info("Stopped bot")
}
