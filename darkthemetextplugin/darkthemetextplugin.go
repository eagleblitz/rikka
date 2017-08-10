package darkthemetextplugin

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"

	"github.com/ThyLeader/rikka"
	"github.com/fogleman/gg"
)

func message(bot *rikka.Bot, service rikka.Service, message rikka.Message) {
	if service.IsMe(message) {
		return
	}

	if !rikka.MatchesCommand(service, "darkonly", message) {
		return
	}

	t, parts := rikka.ParseCommand(service, message)
	if len(parts) == 0 {
		service.SendMessage(message.Channel(), "Please enter some text")
		return
	}

	i := darkThemeGen(t)
	if i == nil {
		service.SendMessage(message.Channel(), "Error generating image")
	}
	service.SendFile(message.Channel(), "dark.png", i)
}

func loadFunc(bot *rikka.Bot, service rikka.Service, data []byte) error {
	return nil
}

func helpFunc(bot *rikka.Bot, service rikka.Service, message rikka.Message, detailed bool) []string {
	if detailed {
		return nil
	}
	return rikka.CommandHelp(service, "darkonly", "text", "Sends an image that is only readable on dark theme.")
}

// New creates a new discordavatar plugin.
func New() rikka.Plugin {
	p := rikka.NewSimplePlugin("DarkThemeText")
	p.LoadFunc = loadFunc
	p.MessageFunc = message
	p.HelpFunc = helpFunc
	return p
}

func darkThemeGen(text string) *io.PipeReader {
	rd, wr := io.Pipe()
	go func() {
		png.Encode(wr, createDoubleTextImage(text, "You need dark theme to view this message", 0xffffff, 0x36393E))
		wr.Close()
	}()
	return rd
}

func createDoubleTextImage(text, text2 string, clr1, clr2 int) image.Image {
	darktheme := createTextImage(text, clr1)
	whitetheme := createTextImage(text2, clr2)

	var highestX int
	var highestY int
	if darktheme.Bounds().Dx() > whitetheme.Bounds().Dx() {
		highestX = darktheme.Bounds().Dx()
	} else {
		highestX = whitetheme.Bounds().Dx()
	}
	if darktheme.Bounds().Dy() > whitetheme.Bounds().Dy() {
		highestY = darktheme.Bounds().Dy()
	} else {
		highestY = whitetheme.Bounds().Dy()
	}

	maskDark := image.NewRGBA(image.Rect(0, 0, highestX, highestY))
	for y := 0; y < highestY; y++ {
		for x := 0; x < highestX; x++ {
			if y%2 == 0 {
				maskDark.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		}
	}

	maskLight := image.NewRGBA(image.Rect(0, 0, highestX, highestY))
	for y := 0; y < highestY; y++ {
		for x := 0; x < highestX; x++ {
			if y%2 != 0 {
				maskLight.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		}
	}

	combined := image.NewRGBA(maskDark.Bounds())
	draw.DrawMask(combined, combined.Bounds(), whitetheme, image.ZP, maskLight, image.ZP, draw.Over)
	draw.DrawMask(combined, combined.Bounds(), darktheme, image.ZP, maskDark, image.ZP, draw.Over)

	return combined
}

func createTextImage(text string, clr int) image.Image {
	textfont, err := gg.LoadFontFace(`fonts/swanse.ttf`, 50)
	if err != nil {
		return nil
	}

	r := (clr >> 16) & 0xff
	g := (clr >> 8) & 0xff
	b := clr & 0xff

	c := gg.NewContext(0, 0)
	c.SetFontFace(textfont)
	var totalHeight float64
	var totalWidth float64
	lines := c.WordWrap(text, 25*10)
	heights := []float64{}

	for _, l := range lines {
		w, h := c.MeasureString(l)
		w *= 1.3
		h *= 1.3
		totalHeight += h
		if w > totalWidth {
			totalWidth = w
		}
		heights = append(heights, h)
	}

	c = gg.NewContext(int(totalWidth), int(totalHeight))
	c.SetFontFace(textfont)
	c.SetColor(color.RGBA{uint8(r), uint8(g), uint8(b), 255})
	for i := 0; i < len(lines); i++ {
		c.DrawStringAnchored(lines[i], 0, float64(i)*heights[i], 0, 1)
	}

	return c.Image()
}
