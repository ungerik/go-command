package main

import (
	"net/http"

	"github.com/domonda/golog/log"

	"github.com/ungerik/go-command"
	"github.com/ungerik/go-command/htmlform"
	"github.com/ungerik/go-fs"
	"github.com/ungerik/go-httpx/httperr"
	"github.com/ungerik/go-httpx/respond"
)

func main() {
	httperr.DebugShowInternalErrorsInResponse = true

	handler, err := htmlform.NewHandler(Example, &ExampleArgs, "Example Form", respond.StaticHTML("<h1>Success!</h1>"))
	if err != nil {
		log.FatalAndPanic(err)
	}

	handler.SetArgDefaultValue("aBool", true)
	handler.SetArgDefaultValue("anInt", 666)
	handler.SetArgDefaultValue("aFloat", 3.1415)

	handler.SetArgOptions(
		"color",
		[]htmlform.Option{
			{Label: "Red", Value: ColorRed},
			{Label: "Green", Value: ColorGreen},
			{Label: "Blue", Value: ColorBlue},
		},
	)
	handler.SetArgDefaultValue("color", ColorGreen)

	log.Info("Listening on http://localhost:8080").Log()
	err = http.ListenAndServe(":8080", handler)
	if err != nil {
		log.FatalAndPanic(err)
	}
}

type Color int

const (
	ColorRed = iota
	ColorGreen
	ColorBlue
)

var ExampleArgs struct {
	command.ArgsDef

	Bool  bool          `arg:"aBool"  desc:"A bool"`
	Int   int           `arg:"anInt"  desc:"An integer"`
	Float float64       `arg:"aFloat" desc:"A float"`
	Color Color         `arg:"color"   desc:"Select a color"`
	File  fs.FileReader `arg:"file"   desc:"Upload file"`

	// Date  date.Date     `arg:"aDate"   desc:"A date"`
	// Time  time.Time     `arg:"aTime"   desc:"A date and time"`
}

func Example(aBool bool, anInt int, aFloat float64, color Color, file fs.FileReader /*, aDate date.Date, aTime time.Time*/) error {
	log.Info("Example").
		Bool("aBool", aBool).
		Int("anInt", anInt).
		Float("aFloat", aFloat).
		Any("color", color).
		Str("file", file.Name()).
		// Str("aDate", string(aDate)).
		// Time("aTime", aTime).
		Log()

	return nil
}
