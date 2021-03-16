package vcfreader

import (
	"bufio"
	"compress/gzip"
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

	defer file.Close()
	defer gz.Close()
	return VCFParser(gz)

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

func VCFParser(file io.Reader) chan []string {
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
			continue

		}
	}()

	return ch
}
