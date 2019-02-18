package main

import (
	"fmt"
	"os"

	"github.com/dnoberon/gifaway/split"

	"github.com/spf13/cobra"
)

// we're going to keep this as a string that we know we can easily parse. Helps with the flags
var splitDuration float64
var outputDirectoryName string

func main() {
	Execute()
}

var rootCmd = &cobra.Command{
	Use:   "gifaway [video file]",
	Short: "gifaway splits video files into gifs with minimal fuss",
	Long:  `gifaway is a simple wrapper of the ffmpeg and ffprobe tool. It was built to make splitting a video file into gifs easy and intuitive.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// cobra will error out if no arg is given, so we _should_ be safe accessing index
		split.Execute(args[0], outputDirectoryName, splitDuration)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().Float64Var(&splitDuration, "duration", 0, "Duration of split gifs in seconds")
	rootCmd.Flags().StringVar(&outputDirectoryName, "output folder name", "results", "Gif output directory name")
}
