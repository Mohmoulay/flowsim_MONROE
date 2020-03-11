package http3

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	// "crypto/tls"
	// "crypto/x509"
	// "fmt"
	quic "github.com/lucas-clemente/quic-go"
	quictrace "github.com/lucas-clemente/quic-go/quictrace"

	http3 "github.com/lucas-clemente/quic-go/http3"
	common "github.com/paaguti/flowsim/common"

	// "io/ioutil"
	"log"
	"net"
	"net/http"

	// "net/url"
	"path"
	"strconv"
	"time"
)

// quicConf := &quic.Config{}
// 	if *trace {
// 		quicConf.QuicTracer = tracer
// 	}
// 	if *qlog {
// 		quicConf.GetLogWriter = func(connID []byte) io.WriteCloser {
// 			filename := fmt.Sprintf("server_%x.qlog", connID)
// 			f, err := os.Create(filename)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			log.Printf("Creating qlog file %s.\n", filename)
// 			return f
// 		}
// 	}
var tracer quictrace.Tracer

func init() {
	tracer = quictrace.NewTracer()
}

func Server(ip string, port int, single bool, tos int, certs string) {
	// var tracer quictrace.Tracer
	trace := flag.Bool("trace", false, "enable quic-trace")
	qlog := flag.Bool("qlog", true, "output a qlog (in the same directory)")

	srvClosed := make(chan int)
	mux := common.MakeHTTPMux(single, srvClosed)
	quicConf := &quic.Config{}
	if *trace {
		quicConf.QuicTracer = tracer
	}
	if *qlog {
		quicConf.GetLogWriter = func(connID []byte) io.WriteCloser {
			// log.Println(connID)
			filename := fmt.Sprintf("HTTP3_%x.qlog", connID)
			f, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Creating qlog file %s.\n", filename)
			return f
		}
	}
	// tracer := quicConf.QuicTracer

	for {
		var server http3.Server

		go func() {

			bCap := net.JoinHostPort(ip, strconv.Itoa(port))
			// connID := quicConf.ConnectionIDLength
			// log.Println(connID)

			log.Printf("Starting HTTP3 server at %s", bCap)
			server = http3.Server{
				Server: &http.Server{
					Addr:           bCap,
					Handler:        mux,
					ReadTimeout:    10 * time.Second,
					WriteTimeout:   10 * time.Second,
					MaxHeaderBytes: 1 << 20,
				},
				//  QuicConfig: &quic.Config{},
				QuicConfig: quicConf,
			}
			certFile := path.Join(certs, "flowsim-server.crt")
			keyFile := path.Join(certs, "flowsim-server.key")
			err := server.ListenAndServeTLS(certFile, keyFile)

			// log.Println(state)
			if err != nil {
				log.Printf(" From http3 server: %v", err)

				// common.Expor
			}
		}()
		<-srvClosed
		if single {
			server.Shutdown(context.Background())
			log.Printf("http3 server shutdown")
			// log.Printf(state)
			break
		}
	}
}
