package main

import (
	"os"
	"waifuIM/cmd/waifu-im/commands"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{Use: "waifu-im-client"}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(commands.NewRandomCMD())
	rootCmd.AddCommand(commands.NewTagsCMD())
	rootCmd.AddCommand(commands.NewArtistsCMD())
	rootCmd.AddCommand(commands.NewAlbumsCMD())
	rootCmd.AddCommand(commands.NewAlbumDetailsCMD())
	rootCmd.AddCommand(commands.NewCreateAlbumCMD())
	rootCmd.AddCommand(commands.NewUpdateAlbumCMD())
	rootCmd.AddCommand(commands.NewDeleteAlbumCMD())
}
