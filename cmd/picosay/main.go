package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gordonklaus/portaudio"
	"github.com/tzneal/gopicotts"
)

func main() {
	lang := flag.String("lang", "en-US", fmt.Sprintf("language to use %v", gopicotts.SupportedLanguages()))
	flag.Parse()

	opts := gopicotts.DefaultOptions
	opts.Language = gopicotts.ParseLanguageName(*lang)
	eng, err := gopicotts.NewEngine(opts)
	defer eng.Close()

	if err != nil {
		log.Fatalf("error intializing engine: %s", err)
	}

	portaudio.Initialize()
	defer portaudio.Terminate()

	// pico outputs data in 16000 hz mono
	const outputChannels = 1
	const sampleRate = 16000

	buf := make([]int16, sampleRate/5)
	strm, err := portaudio.OpenDefaultStream(0, outputChannels, sampleRate, 0, buf)
	if err != nil {
		log.Fatalf("error opening audio stream: %s", err)
	}

	bw := bufwriter{buf, strm, 0}
	eng.SetOutput(bw.processSpeechData)
	if err := strm.Start(); err != nil {
		log.Fatalf("error starting audio stream: %s", err)
	}
	for _, a := range flag.Args() {
		eng.SendText(a)
	}
	// needed in case there is no end of sentence on the input
	eng.FlushSendText()
	strm.Write()
	// flush any remaining data
	if err := strm.Stop(); err != nil {
		log.Fatalf("error stopping audio stream: %s", err)
	}
	strm.Close()
	eng.CloseFileOutput()

}

type bufwriter struct {
	output []int16
	stream *portaudio.Stream
	pos    int
}

func (b *bufwriter) processSpeechData(input []int16) {
	rem := len(input)
	offset := 0
	for rem > 0 {
		// copy our input speech data to the portaudio buffer
		n := copy(b.output[b.pos:], input[offset:])
		rem -= n
		offset += n

		b.pos += n
		if n == 0 {
			if err := b.stream.Write(); err != nil {
				fmt.Println(err)
			}
			b.pos = 0
			// portaudio has the data now, so clear our buffer to prevent a
			// flush in main from replaying remaining data.
			for i := 0; i < len(b.output); i++ {
				b.output[i] = 0
			}
		}
	}
}
