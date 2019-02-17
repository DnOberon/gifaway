package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// we're going to keep this as a string that we know we can easily parse. Helps with the flags
var splitDuration float64
var outputDirectoryName string

var rootCmd = &cobra.Command{
	Use:   "gifaway [video file]",
	Short: "gifaway splits video files into gifs with minimal fuss",
	Long:  `gifaway is a simple wrapper of the ffmpeg and ffprobe tool. It was built to make splitting a video file into gifs easy and intuitive.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// cobra will error out if no arg is given, so we _should_ be safe accessing index
		split(args[0], outputDirectoryName, splitDuration)
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

// TODO bad fileext handling
// TODO leave comments about security
// TODO finish concurrency
// TODO better request struct
// TODO clean output directory
func split(videoName, outputDirectory string, duration float64) {
	dir, err := ioutil.ReadDir(outputDirectory)

	// clean directory if it exists
	// in the case of an empty directory we'll get an error - and ignore it since we'll make it later on
	// if the directory is failing to be read for another reason, the ffmpeg output will (hopefully) clarify
	if err != nil {
		// ffmpeg will not create a dir automatically
		err = os.Mkdir(outputDirectory, os.ModeAppend)
	} else {
		for _, d := range dir {
			os.RemoveAll(path.Join([]string{outputDirectory, d.Name()}...))
		}
	}

	var out bytes.Buffer
	var stderr bytes.Buffer

	durationCommand := exec.Command(`ffprobe`,
		`-v`,
		`error`,
		`-show_entries`,
		`format=duration`,
		`-of`,
		`default=noprint_wrappers=1:nokey=1`,
		videoName,
	)

	durationCommand.Stdout = &out
	durationCommand.Stderr = &stderr

	err = durationCommand.Run()
	if err != nil {
		panic(err)
	}

	// TODO cover assumption of good content, discard the bad. Maybe cast first split string to int?
	// we only want the seconds - could we do this with printf - sure but can't
	// really assume platform
	raw := strings.Split(out.String(), ".")

	movieDuration, err := time.ParseDuration(raw[0] + "s")

	// TODO Cleanup Worker pool functions
	// load the tasks up
	var tasks []*Task

	for i := 1; i < int(movieDuration.Seconds()/duration); i++ {
		ripTask(i*int(duration), (i+1)*int(duration), videoName, outputDirectoryName)

		fmt.Println(i * int(duration))
		tasks = append(tasks, NewTask(func() error {
			return ripTask(i*int(duration), (i+1)*int(duration), videoName, outputDirectoryName)
		}))

	}

	// TODO better error handling
	if err != nil {
		panic(err.Error())
	}

}

// TODO work on compression
// TODO work on better function arguments, allowing for some
// TODO work on safer output directory challenge
func ripTask(startSecond, endSecond int, fileName, outputDirectory string) error {
	var out bytes.Buffer
	var stderr bytes.Buffer

	ripCommand := exec.Command(
		`ffmpeg`,
		`-ss`,
		strconv.Itoa(startSecond),
		`-i`,
		fileName,
		`-t`,
		strconv.Itoa(endSecond),
		`-s`,
		`780x439`,
		fmt.Sprintf("%s/split%d_%d.gif", outputDirectory, startSecond, endSecond),
		`-hide_banner`,
	)

	ripCommand.Stdout = &out
	ripCommand.Stderr = &stderr

	ripCommand.Run()
	fmt.Println(stderr.String())
	fmt.Println(out.String())
	return nil
}
