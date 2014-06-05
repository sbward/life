package main

import (
	_ "expvar"
	"flag"
	"log"
	"net/http"
	"os"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()

	// Enable CPU profiling with the cpuprofile flag.
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		log.Println("CPU profiling enabled; writing to", f.Name())
		defer pprof.StopCPUProfile()
	}

	log.Println("Launching Life server on localhost:8080")

	// Add a websocket server for the simulation.
	http.Handle("/game", http.HandlerFunc(DefaultServer.httpHandler))

	// Add a file server for the HTML5 page.
	handler := http.FileServer(http.Dir("./public"))
	http.Handle("/", &LogHandler{handler})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalln(err)
	}
}

type LogHandler struct {
	http.Handler
}

func (h *LogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("\033[2m"+r.RemoteAddr, r.Method, r.URL, "\033[0m")
	h.Handler.ServeHTTP(w, r)
}
