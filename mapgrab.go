package main

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/secsy/goftp"
)

func main() {
	currentTime := time.Now().Unix()

	// Setting variables for the territory
	gangURL := "https://www.legacy-game.net/maps/map1_gang.png?" + fmt.Sprint(currentTime)
	gangFileName := "/home/resyz/projects/go/legacy/" + fmt.Sprint(currentTime) + "_gang.png"

	// Downloading the territory map
	err := downloadFile(gangURL, gangFileName)
	if err != nil {
		log.Fatal(err)
	}

	// Setting the variables for the overlay
	overlayURL := "https://www.legacy-game.net/maps/map1_overlay.png?" + fmt.Sprint(currentTime)
	overlayFileName := "/home/resyz/projects/go/legacy/" + fmt.Sprint(currentTime) + "_overlay.png"

	// Downloading the overlay map
	err = downloadFile(overlayURL, overlayFileName)
	if err != nil {
		log.Fatal(err)
	}

	gangFile, err := os.Open(gangFileName)
	if err != nil {
		log.Fatal(err)
	}

	srcImage, err := png.Decode(gangFile)
	if err != nil {
		log.Fatal(err)
	}

	defer gangFile.Close()

	overlayFile, err := os.Open(overlayFileName)
	if err != nil {
		log.Fatal(err)
	}

	destImage, err := png.Decode(overlayFile)
	if err != nil {
		log.Fatal(err)
	}

	defer overlayFile.Close()

	b := srcImage.Bounds()
	combinedImage := image.NewRGBA(b)

	draw.Draw(combinedImage, b, srcImage, image.ZP, draw.Src)
	draw.Draw(combinedImage, destImage.Bounds(), destImage, image.ZP, draw.Over)

	// Need to check that the images directory exists
	if _, err := os.Stat("/home/resyz/projects/go/legacy/images"); os.IsNotExist(err) {
		os.Mkdir("/home/resyz/projects/go/legacy/images", 0755)
	}

	combined, err := os.Create("/home/resyz/projects/go/legacy/images/" + fmt.Sprint(currentTime) + "_map.png")
	if err != nil {
		log.Fatal(err)
	}

	enc := png.Encoder{
		CompressionLevel: png.BestCompression,
	}

	enc.Encode(combined, combinedImage)
	defer combined.Close()

	cmd := exec.Command("cwebp", "-lossless", "1", "-q", "100", "/home/resyz/projects/go/legacy/images/"+fmt.Sprint(currentTime)+"_map.png", "-o", "/home/resyz/projects/go/legacy/images/"+fmt.Sprint(currentTime)+"_map.webp")
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()

	err = uploadFile("/home/resyz/projects/go/legacy/images/"+fmt.Sprint(currentTime)+"_map.webp", fmt.Sprint(currentTime)+"_map.webp")
	if err != nil {
		log.Fatal(err)
	}

	// Cleanup the separate overlay files which were downloaded earlier
	err = os.Remove(gangFileName)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Remove(overlayFileName)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Remove("/home/resyz/projects/go/legacy/images/"+fmt.Sprint(currentTime)+"_map.png")
	if err != nil {
		log.Fatal(err)
	}
}

// Function to upload a file
func uploadFile(fileName, baseName string) error {
	//Set config
	config := goftp.Config{
		User:               "###",
		Password:           "###",
		ConnectionsPerHost: 10,
		Timeout:            10 * time.Second,
		Logger:             os.Stderr,
	}

	// Create the client object
	client, err := goftp.DialConfig(config, "###")
	if err != nil {
		panic(err)
	}

	year, month, _ := time.Now().Date()

	// We need to check if the remote directory exists
	_, err = client.Stat("./maps/" + fmt.Sprint(year) + "-" + fmt.Sprint(int(month)))
	if err != nil {
		client.Mkdir("./maps/" + fmt.Sprint(year) + "-" + fmt.Sprint(int(month)))
	}

	mapImage, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	err = client.Store("./maps/"+fmt.Sprint(year)+"-"+fmt.Sprint(int(month))+"/"+baseName, mapImage)
	if err != nil {
		panic(err)
	}

	defer mapImage.Close()

	return nil
}

// Function to download a file
func downloadFile(URL, fileName string) error {
	response, err := http.Get(URL)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New("Received non 200 response code")
	}

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}
