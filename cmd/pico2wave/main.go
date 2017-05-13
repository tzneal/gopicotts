package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tzneal/gopicotts"
)

func main() {
	output := flag.String("o", "", "file to write output to, extension of .wav implies WAVE, .au implies AU, otherwise output is headerless")
	lang := flag.String("lang", "en-US", fmt.Sprintf("language to use %v", gopicotts.SupportedLanguages()))
	flag.Parse()

	if *output == "" {
		flag.Usage()
		os.Exit(1)
	}

	opts := gopicotts.DefaultOptions
	opts.Language = gopicotts.ParseLanguageName(*lang)
	eng, err := gopicotts.NewEngine(opts)
	defer eng.Close()

	if err != nil {
		log.Fatalf("error intializing engine: %s", err)
	}

	eng.SetFileOutput(*output)
	for _, a := range flag.Args() {
		eng.SendText(a)
	}
	// needed in case there is no end of sentence on the input
	eng.FlushSendText()
	eng.CloseFileOutput()

}
