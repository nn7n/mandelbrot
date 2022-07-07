package main

import (
	"image"
	"image/color"
	"math"
	"runtime"
	"sync"
)

func render(xCenter, yCenter, width float64) *image.NRGBA {
	frame := image.NewNRGBA(image.Rect(0, 0, imageX, imageY))
	ratio := float64(imageY) / float64(imageX)

	tasks := make(chan int)
	wg := new(sync.WaitGroup)
	for c := 0; c < runtime.NumCPU(); c++ {
		wg.Add(1)
		go func() {
			for y := range tasks {
				y0 := (float64(y-imageY/2)/float64(imageY))*width*ratio - yCenter
				for x := 0; x < imageX; x++ {
					x0 := (float64(x-imageX/2)/float64(imageX))*width + xCenter
					frame.Set(x, y, colorize(iterate(x0, y0)))
				}
			}
			wg.Done()
		}()
	}
	for line := 0; line < imageY; line++ {
		tasks <- line
	}
	close(tasks)
	wg.Wait()

	return frame
}

func iterate(x0, y0 float64) int {
	var x, xx, y, yy float64
	var i int
	for i = 0; (xx+yy <= 4) && (i < iterations); i++ {
		xx, yy = x*x, y*y
		//y = 2*x*y + y0
		y = math.FMA(2*x, y, y0)
		x = xx - yy + x0
	}
	return i
}

func colorize(i int) color.Color {
	c := color.NRGBA{A: 0xff}
	if i == iterations {
		return c
	}

	hue := 6 * float64(i%0xff) / float64(iterations%0xff)
	off := uint8(0xff * math.Remainder(hue, 1))
	switch int(math.Remainder(hue, 6)) {
	case 0:
		c.R, c.G = 0xff, off
	case 1:
		c.R, c.G = 0xff-off, 0xff
	case 2:
		c.G, c.B = 0xff, off
	case 3:
		c.G, c.B = 0xff-off, 0xff
	case 4:
		c.R, c.B = off, 0xff
	default:
		c.R, c.B = 0xff, 0xff-off
	}
	return c
}
