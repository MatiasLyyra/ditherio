package main

import (
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"os"

	"github.com/MatiasLyyra/ditherio/dither"
)

func main() {
	origFileName := "assets/shiba.png"
	newFileName := "assets/shiba_v2.png"
	origImgFile, err := os.OpenFile(origFileName, os.O_RDONLY, 0777)
	if err != nil {
		log.Fatalf("Failed to open %s error: %s\n", origFileName, err)
	}
	defer origImgFile.Close()
	origImg, _, err := image.Decode(origImgFile)
	if err != nil {
		log.Fatalf("Failed to decode image %s error: %s\n", origFileName, err)
	}

	newImg := dither.Dither(origImg, dither.Burkes, dither.BlackAndWhitePalette)
	newImgFile, err := os.OpenFile(newFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0775)
	if err != nil {
		log.Fatalf("Failed to create image %s error: %s\n", newFileName, err)
	}
	defer newImgFile.Close()
	err = png.Encode(newImgFile, newImg)
	if err != nil {
		log.Fatalf("Image encoding failed %s\n", err)
	}
}
