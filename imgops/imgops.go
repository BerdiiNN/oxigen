package imgops

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"path"
	"strings"

	_ "golang.org/x/image/webp"

	"github.com/fogleman/gg"
	"go.kyoto.codes/zen/v3/httpx"
	"go.kyoto.codes/zen/v3/logic"
	"go.kyoto.codes/zen/v3/mathx"
)

type Text struct {
	Text  string
	Font  string
	Size  float64
	Align gg.Align
	Color color.Color

	Width float64
}

type Point struct {
	X float64
	Y float64
}

// GetRemote downloads remote image,
// saves it into temporary file,
// loads into image object
// and returns that object
// with cleanup function and error.
func GetRemote(href string) (image.Image, func(), error) {
	// Fetch image
	res := httpx.Request("GET", href).Do()
	if res == nil || res.Body == nil {
		return nil, nil, fmt.Errorf("failed to get a valid response from %s", href)
	}
	// Check for HTTP status code
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	
	// Optionally, check Content-Type header
	contentType := res.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return nil, nil, fmt.Errorf("URL did not return an image; got %s", contentType)
	}

	// Decode into bytes
	resbts, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}
	res.Body.Close()
	// Create temp file
	resfile, err := os.CreateTemp("/tmp", "*.oxigen.tmp")
	if err != nil {
		return nil, nil, err
	}
	// Initialize cleanup function
	cleanup := func() {
		os.Remove(resfile.Name())
	}
	// Write bytes into temp file
	if _, err := resfile.Write(resbts); err != nil {
		return nil, cleanup, err
	}
	// Close file
	if err := resfile.Close(); err != nil {
		return nil, cleanup, err
	}
	// Load into object
	obj, err := gg.LoadImage(resfile.Name())
	if err != nil {
		return nil, cleanup, err
	}
	// Return
	return obj, cleanup, nil
}


// RenderText is a method to render a given text
// with parameters on given context and location.
// Also, provides reasonable defaults for this particular project.
func RenderText(img *gg.Context, location Point, text Text) {
	// Defaults
	text.Font = logic.Or(text.Font, "OpenSans-Regular.ttf")
	text.Size = logic.Or(text.Size, 50)
	if mathx.Sum(text.Color.RGBA()) == 0 { // If everything is zero, it means no color provided
		text.Color = color.Black
	}
	// Load font
	if err := img.LoadFontFace(path.Join("assets/dist/fonts", text.Font), text.Size); err != nil {
		panic("error while reading font file")
	}
	// Set color
	img.SetColor(text.Color)
	// Draw
	img.DrawStringWrapped(
		text.Text,
		location.X,
		location.Y,
		0, 0,
		text.Width,
		1.5,
		text.Align,
	)
}
