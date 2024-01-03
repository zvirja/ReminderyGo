package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	tele "gopkg.in/telebot.v3"
)

func main() {
	interruptSignal := make(chan os.Signal)
	signal.Notify(interruptSignal, os.Interrupt)

	botToken, ok := os.LookupEnv("BOT_TOKEN")
	if !ok {
		log.Fatal("BOT_TOKEN env variable is not set")
	}

	pref := tele.Settings{
		Token:       botToken,
		Poller:      &tele.LongPoller{},
		Synchronous: true,
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/run", func(c tele.Context) error {
		msg, err := b.Send(c.Chat(), "Prepare...")
		if err != nil {
			return err
		}

		for i := 0; i < 5; i++ {
			time.Sleep(time.Second)
			msg, err = b.Edit(msg, fmt.Sprintf("Updated. Iter: %v, Now: %v", i+1, time.Now().Format(time.RFC3339)))
			if err != nil {
				return err
			}
		}

		msg, err = b.Edit(msg, fmt.Sprintf("Done!"))
		if err != nil {
			return err
		}

		return nil
	})

	fmt.Println("Starting server..")
	go func() {
		<-interruptSignal
		b.Stop()
	}()

	b.Start()
	fmt.Println("Stopped!")
}
