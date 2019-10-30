package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/shenwei356/bio/seq"
	"github.com/shenwei356/bio/seqio/fastx"
)

var (
	indir     = flag.String("i", ".", "directory with files to filter")
	outdir    = flag.String("o", "./filtered", "directory where to save filtered alignments")
	empty     = flag.Int("e", 0, "maximum empty sequences to allow alignment")
	threshold = flag.Float64("t", 0.0, "fraction of gap or N symbols in sequence to mark it as non-empty")
	gapped    = "N-"
)

func IsEmpty(r *fastx.Record, f float64) bool {
	sf := r.Seq.BaseContent(gapped)
	return sf > f
}

func CheckFile(filename string, e int, f float64) (bool, error) {
	reader, err := fastx.NewReader(seq.DNAredundant, filename, "")
	if err != nil {
		return false, err
	}
	defer reader.Close()

	empty := 0

	for {
		r, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, err
		}

		if IsEmpty(r, f) {
			empty++
		}
	}

	return empty < e, nil
}

func CopyFiles(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func main() {
	flag.Parse()

	files, err := ioutil.ReadDir(*indir)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		filename := file.Name()
		abspath := path.Join(*indir, filename)

		check, err := CheckFile(abspath, *empty, *threshold)
		if err != nil {
			panic(err)
		}

		if check {
			fmt.Printf("File %s: PASS\n", filename)
			newname := path.Join(*outdir, filename)
			if _, err := CopyFiles(abspath, newname); err != nil {
				panic(err)
			}
		} else {
			fmt.Printf("File %s: FAIL\n", filename)
			newname := path.Join(*outdir, "_"+filename)
			if _, err := CopyFiles(abspath, newname); err != nil {
				panic(err)
			}
		}
	}
}
