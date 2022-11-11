package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var version = "dev"

var (
	imageX      int
	imageY      int
	fontDPI     int
	fontFile    string
	textToWrite string
	outfile     string

	printVersion bool
)

func main() {
	flag.BoolVar(&printVersion, "version", false, "print current version")
	flag.IntVar(&imageX, "x", 199, "max width of image")
	flag.IntVar(&imageY, "y", 80, "max height of image")
	flag.IntVar(&fontDPI, "dpi", 184, "font dpi")
	flag.StringVar(&fontFile, "font", "", "font file")
	flag.StringVar(&textToWrite, "text", "hello", "text to print")
	flag.StringVar(&outfile, "o", "out.png", "output file")

	flag.Parse()

	if printVersion {
		fmt.Println(version)

		return
	}

	if err := run(); err != nil {
		fmt.Println(err.Error())
		flag.Usage()
		os.Exit(1)
	}
}

func run() error {
	if fontFile == "" {
		return fmt.Errorf("you need to specify font file")
	}

	fontToWrite, err := readFont(fontFile)
	if err != nil {
		return fmt.Errorf("read font: %w", err)
	}

	newImg, err := fitTextToImg(imageX, imageY, textToWrite, fontToWrite)
	if err != nil {
		return fmt.Errorf("fit text to img: %w", err)
	}

	file, err := os.Create(outfile)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}

	defer file.Close()

	return png.Encode(file, newImg)
}

func fitTextToImg(x int, y int, text string, fnt *truetype.Font) (image.Image, error) {
	var prevImg image.Image

	fontSize := 1
	padding := 5

	for {
		i, p, err := writeText(x, y, text, fnt, float64(fontSize))
		if err != nil {
			return nil, err
		}

		if p.X.Round()+padding > x || p.Y.Round()+padding > y {
			if prevImg == nil {
				return nil, errors.New("can't fit text to image due to size")
			}

			return prevImg, nil
		}

		prevImg = i

		fontSize++
	}
}

func writeText(x int, y int, text string, fnt *truetype.Font, fntSize float64) (image.Image, *fixed.Point26_6, error) {
	img := image.NewGray(image.Rect(0, 0, x, y))

	draw.Draw(img, img.Bounds(), image.White, image.Point{X: 0, Y: 0}, draw.Src)

	fg := image.Black
	c := freetype.NewContext()
	c.SetDPI(float64(fontDPI))
	c.SetFont(fnt)
	c.SetFontSize(fntSize)
	c.SetClip(img.Bounds())
	c.SetSrc(fg)
	c.SetHinting(font.HintingNone)

	c.SetDst(img)
	pt := freetype.Pt(0, 0+int(c.PointToFixed(fntSize)>>6))

	p, err := c.DrawString(text, pt)
	if err != nil {
		return nil, nil, fmt.Errorf("draw string: %w", err)
	}

	return img, &p, nil
}

func readFont(path string) (*truetype.Font, error) {
	fontBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return freetype.ParseFont(fontBytes)
}
