package split

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

	"github.com/apex/log"
)

func Execute(videoName, outputDirectory string, splitTime float64) {
	dir, err := ioutil.ReadDir(outputDirectory)

	// clean directory if it exists
	// in the case of an empty directory we'll get an error - and ignore it since we'll make it later on
	// if the directory is failing to be read for another reason, the ffmpeg output will (hopefully) clarify
	if err != nil {
		err = os.Mkdir(outputDirectory, 0777)
	} else {
		for _, d := range dir {
			if err := os.RemoveAll(path.Join([]string{outputDirectory, d.Name()}...)); err != nil {
				log.Errorf("issue cleaning result directory", err.Error())
			}
		}
	}

	// movie duration in seconds
	movieDuration, err := findDuration(videoName)
	if err != nil {
		log.Errorf("issue finding duration of input file", err.Error())
		return
	}

	var tasks []*Task

	for i := 0; i < int(movieDuration.Seconds()/splitTime); i++ {
		newRip := ripTask(i*int(splitTime), (i+1)*int(splitTime), videoName, outputDirectory)

		newTask := NewTask(i, fmt.Sprintf("starting-%ds", i*int(splitTime)), newRip)

		tasks = append(tasks, newTask)
	}

	// run tasks concurrently using worker pool
	p := NewPool(tasks, 10)
	p.Run()

	for _, completedTask := range tasks {
		if len(completedTask.ErrorBag) > 0 {
			for _, err := range completedTask.ErrorBag {
				log.Errorf("error while processing job - ID: %d, error: %s", completedTask.ID, err.Error())
			}

			continue
		}

		log.Infof("job ID: %d processed successfully", completedTask.ID)
	}
}

func ripTask(startSecond, endSecond int, fileName, outputDirectory string) *exec.Cmd {
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
		fmt.Sprintf("%s/split%d_%d.gif", outputDirectory, startSecond, endSecond),
		`-hide_banner`,
	)

	ripCommand.Stdout = &out
	ripCommand.Stderr = &stderr

	return ripCommand
}

func findDuration(videoName string) (time.Duration, error) {
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

	err := durationCommand.Run()
	if err != nil {
		log.Error(stderr.String())

		return 0, err
	}

	if out.String() != "" {
		log.Info(out.String())
	}

	// TODO cover assumption of good content, discard the bad. Maybe cast first split string to int?
	// we only want the seconds - could we do this with printf - sure but can't
	// really assume platform
	raw := strings.Split(out.String(), ".")

	return time.ParseDuration(raw[0] + "s")
}
