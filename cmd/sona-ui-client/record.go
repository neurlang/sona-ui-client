package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gen2brain/malgo"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

type RecordedSamples struct {
	capturedSamples []int
	mut             sync.Mutex
	capturedText    string
}

func (r *RecordedSamples) GetText() (str string) {
	r.mut.Lock()
	str = r.capturedText
	r.mut.Unlock()
	return
}

func (r *RecordedSamples) GetLastSamples(n int) (ret []int) {
	r.mut.Lock()
	if n <= len(r.capturedSamples) {
		ret = r.capturedSamples[len(r.capturedSamples)-n:]
	} else {
		ret = make([]int, n, n)
		copy(ret, r.capturedSamples)
	}
	r.mut.Unlock()
	return
}
func (r *RecordedSamples) Clear() {
	r.mut.Lock()
	r.capturedSamples = nil
	r.mut.Unlock()
}

// recordFromMicrophone captures audio from the default microphone and saves it as a WAV file
func (r *RecordedSamples) RecordFromMicrophone(stop chan struct{}) (string, error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "recording-*.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	tmpFile.Close() // We'll reopen it later for writing

	// Initialize malgo
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		// Optional logging callback
	})
	if err != nil {
		return "", fmt.Errorf("failed to initialize context: %v", err)
	}
	defer ctx.Uninit()

	// Configure capture device
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = 16000
	deviceConfig.Alsa.NoMMap = 1 // Fix for some Linux systems

	// Prepare for recording
	r.Clear()

	stopRecording := make(chan struct{})
	recordingDone := make(chan struct{})

	var parity byte
	var consecutive byte
	var sos, cnt, max uint64
	// Callback for captured frames
	onRecvFrames := func(pSample2, pSample []byte, framecount uint32) {
		select {
		case <-stopRecording:
			// Don't append more samples if we're stopping
			return
		default:
			if len(pSample) > 0 {
				r.mut.Lock()
				for i := 0; i < len(pSample); i += 2 {
					if i+1 < len(pSample) {
						// Little-endian int16 to int
						value := int(int16(pSample[i+int(parity)]) | int16(pSample[i+1-int(parity)])<<8)
						r.capturedSamples = append(r.capturedSamples, value)
						sos += uint64(value * value)
						cnt++
					}
				}
				r.mut.Unlock()
				if sos/cnt < 111111 {
					consecutive++
					if consecutive == 100 {
						stop <- struct{}{}
					}
				} else {
					consecutive = 0
				}
				if sos/cnt > max {
					max = sos / cnt
				}
			}
			parity ^= byte(len(pSample) & 1)
		}
	}

	// Create and start capture device
	captureCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, captureCallbacks)
	if err != nil {
		return "", fmt.Errorf("failed to initialize device: %v", err)
	}
	defer device.Uninit()

	err = device.Start()
	if err != nil {
		return "", fmt.Errorf("failed to start device: %v", err)
	}

	// Recording has started
	var captured []int

	// Wait for Enter in a goroutine
	go func() {
		<-stop
		close(stopRecording)

		// Give a moment for the last samples to be processed
		time.Sleep(100 * time.Millisecond)
		device.Stop()

		r.mut.Lock()
		captured = r.capturedSamples
		r.mut.Unlock()

		close(recordingDone)
	}()

	// Wait for recording to finish
	<-recordingDone

	// Prepend a second (works better)
	captured = append(make([]int, 16000), captured...)

	// Convert captured samples to WAV
	err = saveAsWAV(tmpFile.Name(), captured, 16000)
	if err != nil {
		return "", fmt.Errorf("failed to save WAV file: %v", err)
	}

	return tmpFile.Name(), nil
}

// saveAsWAV converts raw PCM samples to a proper WAV file
func saveAsWAV(filename string, samples []int, sampleRate int) error {

	// Create audio buffer
	audioBuf := &audio.IntBuffer{
		Data: samples,
		Format: &audio.Format{
			SampleRate:  sampleRate,
			NumChannels: 1,
		},
	}

	// Create and write WAV file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := wav.NewEncoder(file, sampleRate, 16, 1, 1)
	defer encoder.Close()

	return encoder.Write(audioBuf)
}

func (r *RecordedSamples) Run(helptext, host, port string, filePath string, forever bool, start, stop chan struct{}, hangup func(), textCallback func(string)) {

	r.mut.Lock()
	r.capturedText = helptext
	r.mut.Unlock()
again:

	// If no file specified, record from microphone
	if filePath == "" {
		// Wait for Enter to start
		<-start

		// Record from microphone
		recordedFile, err := r.RecordFromMicrophone(stop)
		if err != nil {
			return
		}

		r.Clear()

		hangup()

		filePath = recordedFile
		//fmt.Printf("Recording saved to temp file: %s\n", recordedFile)
	}

	go func(filePath string) {
		r.mut.Lock()
		r.capturedText = "Transcribing..."
		r.mut.Unlock()

		// Construct the base URL with configurable host and port
		baseURL := fmt.Sprintf("http://%s:%s", host, port)
		apiEndpoint := "/v1/audio/transcriptions"
		fullURL := baseURL + apiEndpoint

		// Open the file
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return
		}
		defer file.Close()

		// Create a buffer to write our multipart form
		var requestBody bytes.Buffer
		multipartWriter := multipart.NewWriter(&requestBody)

		// Add the file to the multipart form
		fileWriter, err := multipartWriter.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			fmt.Printf("Error creating form file: %v\n", err)
			return
		}

		// Copy the file content to the multipart section
		_, err = io.Copy(fileWriter, file)
		if err != nil {
			fmt.Printf("Error copying file content: %v\n", err)
			return
		}

		// Close the multipart writer to set the terminating boundary
		multipartWriter.Close()

		// Create the HTTP request
		req, err := http.NewRequest("POST", fullURL, &requestBody)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}

		// Set the content type header to multipart form data with boundary
		req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

		// Send the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error sending request: %v\n", err)
			return
		}
		defer resp.Body.Close()

		// Read the response
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			return
		}

		// Check if the request was successful
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error: API returned status %s\n", resp.Status)
			fmt.Printf("Response: %s\n", string(responseBody))
			return
		}

		// Parse the JSON response to extract the text field
		var result map[string]interface{}
		err = json.Unmarshal(responseBody, &result)
		if err != nil {
			fmt.Printf("Error parsing JSON response: %v\n", err)
			fmt.Printf("Raw response: %s\n", string(responseBody))
			return
		}

		// Extract and print only the text field
		if text, ok := result["text"]; ok {
			r.mut.Lock()
			txt := strings.TrimSpace(text.(string))
			r.capturedText = txt
			println(txt)
			r.mut.Unlock()

			textCallback(txt)

		} else {
			fmt.Println("Error: 'text' field not found in response")
			fmt.Printf("Full response: %s\n", string(responseBody))
		}

		os.Remove(filePath) // Clean up temp file

	}(filePath)

	if forever {
		filePath = ""
		goto again
	}
}
