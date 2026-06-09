// Mkfiles generates binary files of a given size with distinct content, following the name
// pattern f<n>.bin for increasing n.  Try -h for options.
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
)

var (
	numFiles = flag.Int("n", 10, "How many files")
	fileSize = flag.Int("z", 8, "File size in bytes, divisible by 4")
	first    = flag.Int("f", 0, "First number for output file")
)

// Obviously this could run multiple goroutines but I don't know if it would matter much,
// mostly this should be stuck on the file system.

func main() {
	flag.Parse()
	if *numFiles < 0 || *fileSize < 0 || *first < 0 {
		log.Fatal("Negative argument")
	}
	if *fileSize%4 != 0 {
		log.Fatal("File size must divisible by 4")
	}
	for fno := range *numFiles {
		x := rand.Uint32()
		buf := make([]byte, 0, *fileSize)
		for range *fileSize / 4 {
			buf = append(buf, byte(x), byte(x/256), byte(x/(256*256)), byte(x/(256*256*256)))
			x++
		}
		f, err := os.Create(fmt.Sprintf("f%d.bin", fno+*first))
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.Write(buf)
		if err != nil {
			log.Fatal("err")
		}
		f.Close()
	}
}
