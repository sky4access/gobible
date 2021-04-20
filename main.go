package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sky4access/gobible/internal/bible"
)

type Content struct {
	Book    int32  `db:"book"`
	Chapter int32  `db:"chapter"`
	Verse   int32  `db:"verse"`
	Content string `db:"content"`
}

var (
	inputfile string
	lang      string
)

func main() {

	flag.StringVar(&inputfile, "input", "input.yaml", "input yaml file")
	flag.StringVar(&lang, "lang", "eng", "language: eng or kor")
	flag.Parse()

	if _, err := os.Stat(inputfile); os.IsNotExist(err) {
		fmt.Printf("%s does not exist\n", inputfile)
		os.Exit(1)
	}

	if _, err := os.Stat(bible.KRV_FILE); os.IsExist(err) {
		log.Fatalf("%v does not exist\n", bible.KRV_FILE)
	}

	if _, err := os.Stat(bible.ESV_FILE); os.IsExist(err) {
		log.Fatalf("%v does not exist\n", bible.ESV_FILE)
	}

	lang := strings.ToLower(strings.TrimSpace(lang))

	b := bible.NewBileDB(lang, inputfile)

	b.Init()
	b.Fetch()
	fmt.Print(b.Generate())

}
