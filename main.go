package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/rotblauer/catTracks/catTracks"
)

//start the url handlers, special init for everything?

//Toodle to do , Command line port arg, might mover er to main
func main() {
	var porty int
	var clearDBTestes bool
	var testesRun bool
	var buildIndexes bool

	flag.IntVar(&porty, "port", 8080, "port to serve and protect")
	flag.BoolVar(&clearDBTestes, "castrate-first", false, "clear out db of testes prefixed points") //TODO clear only certain values, ie prefixed with testes based on testesRun
	flag.BoolVar(&testesRun, "testes", false, "testes run prefixes name with testes-")              //hope that's your phone's name
	flag.BoolVar(&buildIndexes, "build-indexes", false, "build index buckets for original trackpoints")

	flag.Parse()

	// Open Bolt DB.
	// catTracks.InitBoltDB()
	if bolterr := catTracks.InitBoltDB(); bolterr == nil {
		defer catTracks.GetDB().Close()
	}
	if clearDBTestes {
		e := catTracks.DeleteTestes()
		if e != nil {
			log.Println(e)
		}
	}
	if buildIndexes {
		catTracks.BuildIndexBuckets() //cleverly always returns nil
	}
	// if qterr := catTracks.InitQT(); qterr != nil {
	// 	log.Println("Error initing QT.")
	// 	log.Println(qterr)
	// }
	catTracks.InitMelody()
	catTracks.SetTestes(testesRun) //is false defaulter, false prefixes names with ""

	router := catTracks.NewRouter()

	http.Handle("/", router)

	//go func() {
	//
	//}()

	//go func() {
	//	for {
	//		if e := catTracks.calculateAndStoreStats(7); e != nil {
	//			log.Println(e)
	//		}
	//	}
	//}()

	http.ListenAndServe(":"+strconv.Itoa(porty), nil)
}
