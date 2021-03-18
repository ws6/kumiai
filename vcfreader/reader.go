package vcfreader

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"log"
	"os"
	"sync"

	"io"
	"strings"
)

func ParseFromFileErr(filename string) chan []string {

	file, err := os.Open(filename)

	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer file.Close()

	if !strings.HasSuffix(filename, ".gz") {
		return VCFParser(file)
	}

	gz, err := gzip.NewReader(file)

	if err != nil {
		log.Fatal(err)
		return nil
	}

	defer gz.Close()
	return VCFParser(gz)

}

func ParseFromFile(filename string) chan []string {

	file, err := os.Open(filename)

	if err != nil {
		log.Fatal(err)
		return nil
	}

	gz, err := gzip.NewReader(file)

	if err != nil {
		log.Fatal(err)
		return nil
	}
	done := make(chan struct{})

	go func() {

		select {
		case <-done:

			if err := gz.Close(); err != nil {
				fmt.Println(`err closing gz file`, err.Error())
			}

			if err := file.Close(); err != nil {
				fmt.Println(`err closing file`, err.Error())
			}
			return

		}

	}()
	//
	return VCFParserDone(gz, done)

}

var GetBatchSize, SetBatchSize = func() (
	func() int,
	func(int),
) {

	batchSize := 1000
	var lock = &sync.Mutex{}
	return func() int {
			return batchSize
		},
		func(bs int) {
			lock.Lock()
			batchSize = bs
			lock.Unlock()
		}
}()

func VCFParserDone(file io.Reader, done chan struct{}) chan []string {
	ch := make(chan []string, GetBatchSize())

	scanner := bufio.NewScanner(file)
	go func() {
		defer close(ch)

		for scanner.Scan() {

			line := scanner.Text()

			if strings.HasPrefix(line, `#`) {
				continue
			}

			sp := strings.Split(line, "\t")

			if len(sp) < 8 {
				continue
			}

			ch <- sp

		}
		done <- struct{}{}
	}()

	return ch
}

func VCFParser(file io.Reader) chan []string {

	done := make(chan struct{}, 1)
	return VCFParserDone(file, done)

}

func VCFParserNew(file io.Reader) chan []string {
	ch := make(chan []string, 100)
	buffered := bufio.NewReaderSize(file, 32768*2)

	go func() {
		defer close(ch)
		for {
			line, err := buffered.ReadBytes('\n')
			if err != nil {
				fmt.Println(`readbyte`, err.Error())
				break
			}
			_line := string(line)
			if strings.HasPrefix(_line, "#") {
				continue
			}

			ch <- strings.Split(_line, "\t")
		}

	}()

	return ch
}
