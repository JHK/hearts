// genfavicons renders the Hearts app icon from an SVG source and writes the
// favicon assets used by the web UI. It is the single source of truth for the
// raster outputs; edit the SVG to change the design.
//
// Input:
//
//	--svg  path to the source SVG (default: internal/webui/assets/icon.svg)
//
// Outputs:
//
//	--favicon-ico        32×32  ICO with embedded PNG (transparent background)
//	--apple-touch-icon  180×180 PNG with white background
//
// Run from the repository root:
//
//	go run ./cmd/genfavicons
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func main() {
	svgIn := flag.String("svg", "internal/webui/assets/icon.svg", "source SVG file")
	faviconOut := flag.String("favicon-ico", "internal/webui/assets/favicon.ico", "output path for favicon.ico (32×32)")
	touchIconOut := flag.String("apple-touch-icon", "internal/webui/assets/apple-touch-icon.png", "output path for apple-touch-icon.png (180×180)")
	flag.Parse()

	icon, err := oksvg.ReadIcon(*svgIn, oksvg.WarnErrorMode)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("writing %s\n", *faviconOut)
	must(saveICO(*faviconOut, render(icon, 32, false)))

	fmt.Printf("writing %s\n", *touchIconOut)
	must(savePNG(*touchIconOut, render(icon, 180, true)))
}

// render rasterizes icon at the given pixel size. If whiteBg is true the
// canvas is filled white first (required for apple-touch-icon).
func render(icon *oksvg.SvgIcon, size int, whiteBg bool) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	if whiteBg {
		draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	}
	icon.SetTarget(0, 0, float64(size), float64(size))
	scanner := rasterx.NewScannerGV(size, size, img, img.Bounds())
	icon.Draw(rasterx.NewDasher(size, size, scanner), 1.0)
	return img
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// saveICO writes a minimal ICO file containing a single PNG-encoded image.
func saveICO(path string, img image.Image) error {
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		return err
	}
	pngData := pngBuf.Bytes()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// ICO header
	binary.Write(f, binary.LittleEndian, uint16(0)) // reserved
	binary.Write(f, binary.LittleEndian, uint16(1)) // type: ICO
	binary.Write(f, binary.LittleEndian, uint16(1)) // image count

	// Image directory entry (16 bytes)
	binary.Write(f, binary.LittleEndian, uint8(w&0xFF))        // width  (0 = 256)
	binary.Write(f, binary.LittleEndian, uint8(h&0xFF))        // height
	binary.Write(f, binary.LittleEndian, uint8(0))             // color count
	binary.Write(f, binary.LittleEndian, uint8(0))             // reserved
	binary.Write(f, binary.LittleEndian, uint16(1))            // color planes
	binary.Write(f, binary.LittleEndian, uint16(32))           // bits per pixel
	binary.Write(f, binary.LittleEndian, uint32(len(pngData))) // image data size
	binary.Write(f, binary.LittleEndian, uint32(6+16))         // data offset: header(6) + dir(16)

	_, err = f.Write(pngData)
	return err
}
