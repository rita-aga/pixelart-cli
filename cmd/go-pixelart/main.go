package main

import (
	"fmt"
	"flag"
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

func resizeImage(img image.Image, size int) image.Image {
	// NearestNeighborResampling, BoxResampling, LinearResampling, CubicResampling, LanczosResampling.
	g := gift.New(
		gift.Resize(size, 0, gift.NearestNeighborResampling),
	)

	dst := image.NewRGBA(g.Bounds(img.Bounds()))
	g.Draw(dst, img)

	return dst
}

func applyPalette(img image.Image, palette color.Palette) image.Image {

	return img

}

func diff(a, b float64) float64 {
	if a < b {
	   return b - a
	}
	return a - b
 }

func findClosestColorByHSL(inputColor color.Color, colorMap map[color.Color]bool) color.Color {

	hueImportance := 0.475
	satImportance := 0.2875
	lumImportance := 0.2375

	targetRixel := rgbaToPixel(inputColor.RGBA())
	targetRgb := colorConvert.RGB{
		R: float64(targetRixel.R), G: float64(targetRixel.G), B: float64(targetRixel.B),
	}
	targetHsl := colorConvert.RGB.ToHSL(targetRgb)

	// distanceMap := make(map[color.Color]float64)

	var closestColor color.Color
	var smallestDistance float64

	for c, _ := range colorMap {

		pixel := rgbaToPixel(c.RGBA())
		rgb := colorConvert.RGB{
			R: float64(pixel.R), G: float64(pixel.G), B: float64(pixel.B),
		}
		hsl := colorConvert.RGB.ToHSL(rgb)
		fmt.Printf("Hue: %v, Saturation: %v, Luminance: %v\n", hsl.H, hsl.S, hsl.L)

		hueDiff := diff(hsl.H, targetHsl.H)
		fmt.Printf("hsl: %v, targetHsl: %v, diff: %v\n", hsl.H, targetHsl.H, hueDiff)

		satDiff := diff(hsl.S, targetHsl.S)
		lumDiff := diff(hsl.L, targetHsl.L)

		// get weighted distance from target
		distanceFromTarget := hueDiff * hueImportance + satDiff * satImportance + lumDiff * lumImportance

		if closestColor == nil {
			// save current color as closest
			closestColor = c
			smallestDistance = distanceFromTarget
		} else {
			// compare to previous closest color
			if smallestDistance > distanceFromTarget {
				smallestDistance = distanceFromTarget
			}
		}

	}

	return closestColor

}

func mapColors(srcColors []color.Color, dstColors []color.Color) map[color.Color]color.Color {
	colorMap := make(map[color.Color]color.Color)

	sourceColorSet := make(map[color.Color]bool)

	for _, s := range srcColors {
		sourceColorSet[s] = true
	}

	for _, c := range dstColors {

		matchingSourceColor := findClosestColorByHSL(c, sourceColorSet)
		
		colorMap[matchingSourceColor] = c

		// delete matched color from source so it's only matched once
		delete(sourceColorSet, matchingSourceColor)
	}

	return colorMap
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

// define pixel struct
type Pixel struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

func recolorImgWithColorMap(srcImg image.Image, colorMap map[color.Color]color.Color) image.Image {

	recoloredImg := image.NewRGBA(srcImg.Bounds())
	draw.Draw(recoloredImg, srcImg.Bounds(), srcImg, image.Point{}, draw.Over)

	width := srcImg.Bounds().Max.X
	height := srcImg.Bounds().Max.Y

	for x := 0; x <= width; x++ {
		for y := 0; y <= height; y++ {
			srcColor := srcImg.At(x, y)
			targetColor := colorMap[srcColor]
			// log.Print(x, y, srcColor, targetColor)
			if targetColor != nil {
				recoloredImg.Set(x, y, targetColor)
			}
		}
	}

	return recoloredImg

}

var blockNum int

func main() {

	// fileName := "../../fauna.png"
	// colorNum := 6
	// blockNum := 40

	// set up flag CLI input
	var fileName string
	var colorNum int
	// var blockNum int

    flag.StringVar(&fileName, "filename", "", "File to convert. (Required)")
	flag.StringVar(&fileName, "f", "", "File to convert. (Required)")
	flag.IntVar(&colorNum, "colors", 6, "Number of colors in the output.")
	flag.IntVar(&colorNum, "c", 6, "Number of colors in the output.")
	flag.IntVar(&blockNum, "blocks", 40, "Number of pixels on the WIDER side of the output.")
	flag.IntVar(&blockNum, "b", 40, "Number of pixels on the WIDER side of the output.")

    flag.Parse()

    if fileName == "" {
        flag.PrintDefaults()
        os.Exit(1)
    }

    fmt.Printf("fileName: %s\n", fileName)

	srcImg := openImage(fileName)
	srcImg = resizeImage(srcImg, 400)

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
	writeImage("01-websafe.jpg", webSafeDst)

	// web safe image limited to X colors
	ltdColorsDst := colorquant.NoDither.Quantize(webSafeDst, webSafeDst, colorNum, false, true)
	writeImage("02-websafe-X-colors.jpg", ltdColorsDst)

	// pixelate X color img
	g := gift.New(
		// gift.Brightness(-25),
		gift.Contrast(20),
		gift.Pixelate(pixelateCoeff),
	)

	ltdColorsPixelatedDst := image.NewRGBA(g.Bounds(ltdColorsDst.Bounds()))
	g.Draw(ltdColorsPixelatedDst, ltdColorsDst)
	writeImage("03-websafe-X-colors-pixelated.jpg", ltdColorsPixelatedDst)

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
		color.RGBA{255, 255, 0, 255}, // yellow
		color.RGBA{0, 255, 0, 255}, // green
		color.RGBA{0, 0, 255, 255}, // blue
		color.RGBA{255, 0, 0, 255}, // red
		color.RGBA{255, 128, 0, 255}, // orange
		color.RGBA{255, 255, 255, 255}, // white
	}

	colorMap := mapColors(imageColors, myColors)
	log.Print(colorMap)

	// myPalette := color.Palette(myColors)

	// myColorDst := image.NewPaletted(srcImg.Bounds(), myPalette)
	// draw.Draw(myColorDst, srcImg.Bounds(), ltdColorsDst2, image.ZP, draw.Over)
	// writeImage("08-websafe-X-colors-pixelated-X-colors-my-color.jpg", myColorDst)

	targetImg := recolorImgWithColorMap(ltdColorsDst2, colorMap)
	writeImage("10-target.jpg", targetImg)


	// NOTES
	// first use this to split image into X color evenly distant from each other then sharp pixelate
	// extract palette from recolored-no-dither
	// map colors to colors from my palette
	// recolor pixels if image using mapping
	// smart calc for for contrast filter values
	// experiment with sharpness
	// experiment with adding contrast/brightness/sharpness as first step

}
