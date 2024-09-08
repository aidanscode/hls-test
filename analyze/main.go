package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	args := os.Args
	fileName := args[1]
	rawCommand := fmt.Sprintf("ffprobe -v quiet -show_format -show_streams -print_format json %s", fileName)
	split := strings.Split(rawCommand, " ")
	cmd := exec.Command(split[0], split[1:]...)
	var out strings.Builder
	var errOut strings.Builder

	cmd.Stdout = &out
	cmd.Stderr = &errOut

	err := cmd.Run()
	if err != nil {
		fmt.Println("Error!", err.Error(), errOut.String())
		return
	}

	fmt.Println("Ok!")
	response := out.String()

	var asJson map[string]any
	err = json.Unmarshal([]byte(response), &asJson)
	if err != nil {
		fmt.Println("Error decoding JSON!", err.Error())
		return
	}

	streams := asJson["streams"].([]interface{})
	stream := streams[0].(map[string]any)
	duration := stream["duration"].(string)
	sampleRate := stream["sample_rate"].(string)
	channels := stream["channels"].(float64)
	encoding := stream["codec_name"].(string)
	bitRate := stream["bit_rate"].(string)

	fmt.Printf("Audio file has encoding %s, sample rate %v, %v channel(s), %v bit rate, and is %s seconds long", encoding, sampleRate, channels, bitRate, duration)

}
