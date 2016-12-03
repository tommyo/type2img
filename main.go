package main

import(
	"os"
	"fmt"
	"io/ioutil"
	"errors"
	"image"
	"image/png"
	"image/draw"
	"strconv"

	"github.com/urfave/cli"
	"github.com/golang/freetype"
	"golang.org/x/image/font"
	"github.com/golang/freetype/truetype"
)

func command(c *cli.Context) error {
	if !c.IsSet("font") {
		return errors.New("font is required")
	}
	if !c.IsSet("output") {
		return errors.New("output file is required")
	}
	args := c.Args()
	if len(args) != 1 {
		return errors.New("1 and exactly 1 strings expected")
	}

	s, err := strconv.Unquote(`'` + args[0] + `'`)
	if err != nil {
		return err
	}

	// load font
	fontBytes, err := ioutil.ReadFile(c.String("font"))
	if err != nil {
		return err
	}
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return err
	}

	// set palette
	bg := image.Transparent
	fg := image.Black

	// initialize renderer
	build := freetype.NewContext()
	build.SetFont(f)
	build.SetFontSize(c.Float64("points"))

	if c.Bool("full") {
		build.SetHinting(font.HintingFull)
	}

	face := truetype.NewFace(f, &truetype.Options{ Size: c.Float64("points") })
	metrics := face.Metrics()

	padding := c.Int("pad")

	width := font.MeasureString(face, s).Ceil()
	height := metrics.Ascent.Ceil() + metrics.Descent.Ceil()
	base := metrics.Ascent.Ceil()

	// initialize image
	img := image.NewRGBA(image.Rect(0, 0, padding + width + padding, padding + height + padding))
	draw.Draw(img, img.Bounds(), bg, image.ZP, draw.Src)

	// start process
	build.SetSrc(fg)
	build.SetClip(img.Bounds())
	build.SetDst(img)

	_, err = build.DrawString(s, freetype.Pt(padding, base + padding))
	if err != nil {
		return err
	}

	if c.Bool("n") {
		fmt.Printf("glyph dimensions: %d, %d\n", width, height)
		fmt.Printf("image dimensions: %d, %d\n", img.Rect.Dx(), img.Rect.Dy())
		fmt.Printf("glyph baseline: %d\n", base)
		return nil
	}


	// initialize file
	out, err := os.Create(c.String("output"))
	if err != nil {
		return err
	}
	defer out.Close()
	return png.Encode(out, img)
}



func main() {
	app := cli.NewApp()

	app.Version = "0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "f, font",
			Usage: "Font file `SOURCE`",
		},
		cli.StringFlag{
			Name: "o, output",
			Usage: "`FILENAME` of output",
		},
		cli.Float64Flag{
			Name: "p, points",
			Usage: "Font size in `PTS`",
			Value: 12,
		},
		cli.IntFlag{
			Name: "pad",
			Usage: "Set padding to `INT`",
		},
		cli.BoolFlag{
			Name: "full",
			Usage: "Turn full font hinting on",
		},
		cli.BoolFlag{
			Name: "n",
			Usage: "Display output information, but don't actually generate file",
		},
	}

	app.Action = command

	app.Run(os.Args)
}
