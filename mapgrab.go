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
	"time"
)

func main() {
	currentTime := time.Now().Unix()

	gangURL := "https://www.legacy-game.net/maps/map1_gang.png?" + fmt.Sprint(currentTime)
	gangFileName := "/home/resyz/projects/go/legacy/" + fmt.Sprint(currentTime) + "_gang.png"
	err := downloadFile(gangURL, gangFileName)
	if err != nil {
		log.Fatal(err)
	}

	overlayURL := "https://www.legacy-game.net/maps/map1_overlay.png?" + fmt.Sprint(currentTime)
	overlayFileName := "/home/resyz/projects/go/legacy/" + fmt.Sprint(currentTime) + "_overlay.png"
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

	png.Encode(combined, combinedImage)
	defer combined.Close()

	// Cleanup the separate overlay files which were downloaded earlier
	err = os.Remove(gangFileName)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Remove(overlayFileName)
	if err != nil {
		log.Fatal(err)
	}
}

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
