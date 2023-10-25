package bot

import (
	"log"

	"github.com/AndreyAD1/helsinki-guide/internal/bot"
	"github.com/spf13/cobra"
)

var (
	botToken string
	BotCmd   = &cobra.Command{
		Use:   "bot",
		Short: "Run a Telegram bot",
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}
)

func init() {
	BotCmd.Flags().StringVarP(
		&botToken,
		"token",
		"t",
		"",
		"a token of Telegram bot",
	)
	BotCmd.MarkFlagRequired("token")
}

func run() {
	server, err := bot.NewServer(botToken)
	if err != nil {
		log.Fatalf("can not run a server: %v", err)
	}
	server.RunBot()
}
