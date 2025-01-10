package adapter

import (
	"codebase-app/internal/infrastructure/config"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	tele "gopkg.in/telebot.v3"
)

// WithVenamonGolog sets up a new Venamon Golog bot using the provided token and settings.
//
// It assigns the newly created bot to the Adapter's VenamonGolog field.
// When using this option for hooks or commands, make sure to check if the bot is not nil.
func WithVenamonGolog() Option {
	return func(a *Adapter) {

		bot, err := tele.NewBot(tele.Settings{
			Token:  config.Envs.VenamonGolog.Token,
			Poller: &tele.LongPoller{Timeout: 10 * time.Second},
		})

		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create Venamon Golog")
		}

		_, err = bot.Send(
			&tele.Chat{
				ID:   config.Envs.VenamonGolog.ChatId,
				Type: tele.ChatGroup,
			},
			fmt.Sprintf("Hello, Venamon Golog is running at %v", time.Now().Format(time.DateTime)),
			&tele.SendOptions{
				ThreadID:  config.Envs.VenamonGolog.ThreadId,
				ParseMode: tele.ModeMarkdown,
			})
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to send message to Venamon Golog")
			return
		}

		a.VenamonGolog = bot
		log.Info().Msg("Venamon Golog is running")
	}
}
