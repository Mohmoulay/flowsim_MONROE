package common

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	quictrace "github.com/lucas-clemente/quic-go/quictrace"
)

func getInt(query url.Values, field string) (int, bool) {
	params, ok := query[field]
	if !ok {
		log.Println("Url Param '" + field + "' is missing")
		return -1, false
	}
	value, err := strconv.Atoi(params[0])
	return value, err == nil
}

var state quictrace.TransportState
var tracer quictrace.Tracer

func init() {
	tracer = quictrace.NewTracer()
}

func exportTraces() error {
	traces := tracer.GetAllTraces()
	log.Println(traces)
	if len(traces) != 1 {
		return errors.New("expected exactly one trace")
	}
	for _, trace := range traces {
		f, err := os.Create("/home/mlt/trace.qtr")
		log.Println("saved filed")
		if err != nil {
			return err
		}
		if _, err := f.Write(trace); err != nil {
			return err
		}
		f.Close()
		fmt.Println("Wrote trace to", f.Name())
	}
	return nil
}

// type tracingHandler struct {
// 	handler http.Handler
// }

// var _ http.Handler = &tracingHandler{}

// func (h *tracingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	h.handler.ServeHTTP(w, r)
// 	if err := exportTraces(); err != nil {
// 		log.Fatal(err)
// 	}
// }

func MakeHTTPMux(single bool, srvClosed chan int) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/flowsim/request", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		log.Printf("Received %s%v", r.URL.Path, query)
		requested, ok := getInt(query, "bytes")
		if !ok {
			log.Println("Couldn't decode " + r.URL.Path)
			return
		}
		pass, ok := getInt(query, "pass")
		if !ok {
			log.Println("Couldn't decode " + r.URL.Path)
			return
		}
		total, ok := getInt(query, "of")
		if !ok {
			log.Println("Couldn't decode " + r.URL.Path)
			return
		}

		// fmt.Fprintf(w, "So you are requesting "+bytes+" bytes in pass "+pass+" of "+of+" from me...")
		fmt.Fprintln(w, RandStringBytes(requested))
		if pass == total {
		}
		log.Printf("Served %d bytes in pass %d of %d", requested, pass, total)
	})

	mux.HandleFunc("/flowsim/close", func(w http.ResponseWriter, req *http.Request) {
		// if req.URL.Path != "/" {
		// 	http.NotFound(w, req)
		// 	log.Fatal("Can't serve " + req.URL.Path)
		// } else {
		fmt.Fprintf(w, "Bye!\n")
		// }
		if single {
			log.Printf("And here we should stop")
			srvClosed <- 1
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			log.Fatal("Can't serve " + req.URL.Path)
		} else {
			fmt.Fprintf(w, "Welcome to the flowsim QUIC server!")
		}
	})

	return mux
}

// 	return &tracingHandler{handler: mux}

// }
