package gopicotts

import (
	"log"

	"github.com/tzneal/gopicotts"
)

func ExampleNewEngine() {
	eng, err := gopicotts.NewEngine(gopicotts.DefaultOptions)
	defer eng.Close()
	if err != nil {
		log.Fatalf("error intializing engine: %s", err)
	}
	eng.SetFileOutput("output.wav")
	eng.SendText("hello world.")
	eng.CloseFileOutput()
}
