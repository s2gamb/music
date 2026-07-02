package main

import (
	"fmt"
	"os"

	"github.com/s2gamb/music/internal/cli"
	"github.com/spf13/cobra"
)

var (
	ldTelegramBotToken string
	ldTelegramChatID   string
	ldDatabaseURL      string
)

var rootCmd = &cobra.Command{
	Use:   "music",
	Short: "Music CLI tool for managing albums and tracks",
}

func init() {
	os.Setenv("TELEGRAM_BOT_TOKEN", ldTelegramBotToken)
	os.Setenv("TELEGRAM_CHAT_ID", ldTelegramChatID)
	os.Setenv("DATABASE_URL", ldDatabaseURL)

	rootCmd.AddCommand(cli.UploadCmd)
	rootCmd.AddCommand(cli.EditCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
