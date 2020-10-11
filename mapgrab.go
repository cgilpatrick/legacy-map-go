package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"image"
	"image/draw"
	"image/png"
)

func main() {
	currentTime := time.Now().Unix()
	gangURL := "https://www.legacy-game.net/maps/map1_gang.png?" + fmt.Sprint(currentTime)
	gangFileName := fmt.Sprint(currentTime) + "_gang.png"
	err := downloadFile(gangURL, gangFileName)
	if err != nil {
		log.Fatal(err)
	}
	
	overlayURL := "https://www.legacy-game.net/maps/map1_overlay.png?" + fmt.Sprint(currentTime)
	overlayFileName := fmt.Sprint(currentTime) + "_overlay.png"
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
	
	combined, err := os.Create("result.png")
	if err != nil {
		log.Fatal(err)
	}

	png.Encode(combined, combinedImage)
	defer combined.Close()
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