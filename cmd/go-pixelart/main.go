package main

import (
	"fmt"
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
	colorConvert "github.com/gerow/go-color"
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

func applyPalette(img image.Image, palette color.Palette) image.Image {

	return img

}

func mapColors(srcColors []color.Color, dstColors []color.Color) map[color.Color]color.Color {
	result := make(map[color.Color]color.Color)

	for _, c := range srcColors {
		//
		pixel := rgbaToPixel(c.RGBA())
		rgb := colorConvert.RGB{
			R: float64(pixel.R), G: float64(pixel.G), B: float64(pixel.B),
		}
		hsl := colorConvert.RGB.ToHSL(rgb)
		fmt.Printf("Hue: %v, Saturation: %v, Luminance: %v\n", hsl.H, hsl.S, hsl.L)
	}

	return result
}

func uniqueColors(img image.Image) ([]color.Color, error) {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	var pixels [][]Pixel
	result := make(map[Pixel]int)
	for y := 0; y < height; y++ {
		var row []Pixel
		for x := 0; x < width; x++ {
			result[rgbaToPixel(img.At(x, y).RGBA())] = 1
			row = append(row, rgbaToPixel(img.At(x, y).RGBA()))
		}
		pixels = append(pixels, row)
	}

	uniqueColors := make([]color.Color, 0)
	for k := range result {
		uniqueColors = append(uniqueColors, color.RGBA{k.R, k.G, k.B, k.A})
	}

	return uniqueColors, nil
}

func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
	return Pixel{uint8(r / 257), uint8(g / 257), uint8(b / 257), uint8(a / 257)}
}

// Pixel struct example
type Pixel struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

func main() {
	fileName := "fauna.png"

	colorNum := 6
	blockNum := 40

	srcImg := openImage(fileName)
	srcImg = resizeImage(srcImg)

	width := srcImg.Bounds().Max.X
	height := srcImg.Bounds().Max.Y

	var pixelateCoeff int

	if width > height {
		pixelateCoeff = width / blockNum
	} else {
		pixelateCoeff = height / blockNum
	}

	log.Printf("Width: %v", width)
	log.Printf("Height: %v", height)
	log.Printf("Pixelate Coeff: %v", pixelateCoeff)

	// web safe image
	webSafeDst := image.NewPaletted(srcImg.Bounds(), palette.WebSafe)
	draw.Draw(webSafeDst, srcImg.Bounds(), srcImg, image.ZP, draw.Over)
	// writeImage("01-websafe.jpg", webSafeDst)

	// web safe image limited to X colors
	ltdColorsDst := colorquant.NoDither.Quantize(webSafeDst, webSafeDst, colorNum, false, true)
	// writeImage("02-websafe-X-colors.jpg", ltdColorsDst)

	// pixelate X color img
	g := gift.New(
		// gift.Brightness(-25),
		gift.Contrast(20),
		gift.Pixelate(pixelateCoeff),
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

	imageColors, err := uniqueColors(ltdColorsDst2)
	if err != nil {
		log.Fatalf("Error extracting colors: %v", err)
	}
	log.Printf("Number of unique colors: %v", len(imageColors))
	log.Print(imageColors)

	myColors := []color.Color{
		color.RGBA{84, 19, 136, 255},
		color.RGBA{217, 3, 104, 255},
		color.RGBA{241, 233, 218, 255},
		color.RGBA{46, 41, 78, 255},
		color.RGBA{255, 212, 0, 255},
		color.RGBA{49, 175, 212, 255},
	}

	mapColors(imageColors, myColors)

	// myPalette := color.Palette(myColors)

	// myColorDst := image.NewPaletted(srcImg.Bounds(), myPalette)
	// draw.Draw(myColorDst, srcImg.Bounds(), ltdColorsDst4, image.ZP, draw.Over)
	// writeImage("08-websafe-X-colors-pixelated-X-colors-my-color.jpg", myColorDst)

	// NOTES
	// first use this to split image into X color evenly distant from each other then sharp pixelate
	// extract palette from recolored-no-dither
	// map colors to colors from my palette
	// recolor pixels if image using mapping
	// smart calc for for contrast filter values
	// experiment with sharpness
	// experiment with adding contrast/brightness/sharpness as first step

}
