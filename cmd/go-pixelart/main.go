package main

import (
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"

	"github.com/disintegration/gift"
	"github.com/esimov/colorquant"
)

func writeImage(fileName string, file image.Image) {
	out, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Error creating output file %s: %v", fileName, err)
	}
	defer out.Close()

	if err := jpeg.Encode(out, file, nil); err != nil {
		log.Fatalf("Error outputting image: %v", err)
	}

	log.Printf("Output image: %s", fileName)
}

func openImage(fileName string) image.Image {
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Error loading image: %v", err)
	}
	defer f.Close()

	srcImg, _, err := image.Decode(f)
	if err != nil {
		log.Fatalf("Error decoding file to an Image: %v", err)
	}

	_ = srcImg
	log.Printf("Processing image: %s", fileName)

	return srcImg
}

func resizeImage(img image.Image) image.Image {
	g := gift.New(
		gift.Resize(700, 0, gift.LanczosResampling),
	)

	dst := image.NewRGBA(g.Bounds(img.Bounds()))
	g.Draw(dst, img)

	return dst
}

func main() {
	fileName := "photo.jpg"

	colorNum := 6

	srcImg := openImage(fileName)
	srcImg = resizeImage(srcImg)

	width := srcImg.Bounds().Max.X
	height := srcImg.Bounds().Max.Y

	log.Printf("Width: %v", width)
	log.Printf("Height: %v", height)

	// web safe image
	webSafeDst := image.NewPaletted(srcImg.Bounds(), palette.WebSafe)
	draw.Draw(webSafeDst, srcImg.Bounds(), srcImg, image.ZP, draw.Over)
	// writeImage("01-websafe.jpg", webSafeDst)

	// web safe image limited to X colors
	ltdColorsDst := colorquant.NoDither.Quantize(webSafeDst, webSafeDst, colorNum, false, true)
	// writeImage("02-websafe-X-colors.jpg", ltdColorsDst)

	// pixelate X color img
	g := gift.New(
		// TODO make image smaller first
		// TODO smart calc for for these values
		// TODO experiment with sharpness
		// TODO experiment with adding this in the beginning
		// gift.Brightness(-15),
		gift.Contrast(30),
		gift.Pixelate(width/100),
	)

	ltdColorsPixelatedDst := image.NewRGBA(g.Bounds(ltdColorsDst.Bounds()))
	g.Draw(ltdColorsPixelatedDst, ltdColorsDst)
	// writeImage("03-websafe-X-colors-pixelated.jpg", pm5)

	sourcePixelatedDst := image.NewRGBA(g.Bounds(ltdColorsDst.Bounds()))
	g.Draw(sourcePixelatedDst, srcImg)
	writeImage("04-source-pixelated.jpg", sourcePixelatedDst)

	// web safe image limited to X colors
	ltdColorsDst2 := colorquant.NoDither.Quantize(sourcePixelatedDst, sourcePixelatedDst, colorNum, false, true)
	writeImage("05-websafe-pixelated-X-colors.jpg", ltdColorsDst2)

	// srcImg2 := openImage("tutu_pixelated.jpg")
	// ltdColorsDst3 := colorquant.NoDither.Quantize(srcImg2, webSafeDst, colorNum, false, true)
	// writeImage("06-ext-pixelated-to-ltd-color.jpg", ltdColorsDst3)

	ltdColorsDst4 := colorquant.NoDither.Quantize(ltdColorsPixelatedDst, ltdColorsPixelatedDst, colorNum, false, true)
	writeImage("07-websafe-X-colors-pixelated-X-colors.jpg", ltdColorsDst4)

	colors := []color.Color{
		color.RGBA{84, 19, 136, 255},
		color.RGBA{217, 3, 104, 255},
		color.RGBA{241, 233, 218, 255},
		color.RGBA{46, 41, 78, 255},
		color.RGBA{255, 212, 0, 255},
		color.RGBA{49, 175, 212, 255},
	}

	myPalette := color.Palette(colors)

	myColorDst := image.NewPaletted(srcImg.Bounds(), myPalette)
	draw.Draw(myColorDst, srcImg.Bounds(), ltdColorsDst4, image.ZP, draw.Over)
	writeImage("08-websafe-X-colors-pixelated-X-colors-my-color.jpg", myColorDst)

	// ditherer := colorquant.Dither{
	// 	[][]float32{
	// 		[]float32{0.0, 0.0, 0.0, 7.0 / 48.0, 5.0 / 48.0},
	// 		[]float32{3.0 / 48.0, 5.0 / 48.0, 7.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0},
	// 		[]float32{1.0 / 48.0, 3.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0, 1.0 / 48.0},
	// 	},
	// }
	// pm2 := ditherer.Quantize(img, pm, 6, true, true)
	// draw.Draw(pm2, img.Bounds(), img, image.ZP, draw.Over)

	// [TODO] extract palette from recolored-no-dither
	// [TODO] map colors to colors from my palette
	// [TODO] recolor pixels if image using mapping

}
