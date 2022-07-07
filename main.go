package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"
)

var imageX, imageY, iterations int

func main() {
	var (
		reps                     int
		profiling, verbose, help bool
		xPos, yPos, span, scale  float64
	)
	flag.IntVar(&imageX, "iw", 0, "*image width in `pixels` (X)")
	flag.IntVar(&imageY, "ih", 0, "*image height in `pixels` (Y)")
	flag.Float64Var(&xPos, "x", 0, "center of image (X/Re)")
	flag.Float64Var(&yPos, "y", 0, "center of image (Y/Im)")
	flag.Float64Var(&span, "s", 0, "*X/Re `span` of the image")
	flag.IntVar(&iterations, "i", 0, "*number of `iterations` during computation")
	flag.IntVar(&reps, "r", 1, "number of consecutive output `images`")
	flag.Float64Var(&scale, "z", 0.9, "scaling `factor` of each consecutive image if -r>1 (0.8=80%)")
	flag.BoolVar(&profiling, "p", false, "enable CPU and memory profiling")
	flag.BoolVar(&verbose, "v", false, "print additional information during execution")
	flag.BoolVar(&help, "h", false, "print this message")
	flag.Parse()
	log.SetFlags(0)
	if imageX < 1 || imageY < 1 || iterations < 1 || reps < 1 || span == 0 || scale == 0 || help {
		flag.PrintDefaults()
		log.Println("  * - mandatory values")
		os.Exit(0)
	}

	if profiling {
		cp, err := os.Create("cpuprofile")
		if err != nil {
			log.Fatalln("Could not create CPU profile: ", err)
		}
		defer cp.Close()
		if err = pprof.StartCPUProfile(cp); err != nil {
			log.Fatalln("Could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	ticks := time.Tick(5 * time.Second)
	start := time.Now()
	for i := 0; i < reps; i++ {
		name := strconv.Itoa(i) + ".png"
		file, err := os.Create(name)
		if err != nil {
			log.Fatalln("Error creating file", name, err)
		}

		frame := render(xPos, yPos, span)
		span *= scale

		if err = png.Encode(file, frame); err != nil {
			log.Fatalln("Encoding error, img#", i, err)
		}
		if err = file.Close(); err != nil {
			log.Fatalln("Error closing file", name, err)
		}

		select {
		case t := <-ticks:
			if verbose {
				log.Println(t.Sub(start).Round(time.Second), fmt.Sprintf("\t%d%%\t%d/%d", i*100/reps, i, reps))
			}
		case <-signals:
			log.Println("Interrupted")
			os.Exit(0)
		default:
		}
	}

	if profiling {
		mp, err := os.Create("memprofile")
		if err != nil {
			log.Fatalln("Could not create memory profile: ", err)
		}
		defer mp.Close()
		runtime.GC()
		if err = pprof.WriteHeapProfile(mp); err != nil {
			log.Fatalln("Could not write memory profile: ", err)
		}
	}
}
