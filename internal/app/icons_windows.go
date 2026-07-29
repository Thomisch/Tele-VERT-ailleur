//go:build windows

package app

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// Icônes tray générées au démarrage en PNG (format attendu par Fyne) : un
// disque plein de couleur sur fond transparent. Évite tout fichier externe.
var (
	iconActivePNG = makePNG(0x2E, 0xCC, 0x71) // vert
	iconPausedPNG = makePNG(0x95, 0xA5, 0xA6) // gris
)

const iconSize = 32

// makePNG produit les octets d'un PNG contenant un disque coloré opaque sur
// fond transparent.
func makePNG(r, g, b uint8) []byte {
	const cx, cy, radius = iconSize / 2, iconSize / 2, iconSize/2 - 2
	img := image.NewNRGBA(image.Rect(0, 0, iconSize, iconSize))
	fill := color.NRGBA{R: r, G: g, B: b, A: 0xFF}
	for y := 0; y < iconSize; y++ {
		for x := 0; x < iconSize; x++ {
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy <= radius*radius {
				img.SetNRGBA(x, y, fill)
			}
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
