package quic

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"os"

	// "crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"strconv"
	"time"

	quic "github.com/lucas-clemente/quic-go"

	common "github.com/paaguti/flowsim/common"
)

// var tracer quictrace.Tracer

// Start a server that echos all data on the first stream opened by the client
func Server(ip string, port int, single bool, dscp int) error {
	quicConf := &quic.Config{}
	addr := net.JoinHostPort(ip, strconv.Itoa(port))
	qlog := flag.Bool("qlog", true, "output a qlog (in the same directory)")

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if *qlog {
		quicConf.GetLogWriter = func(connID []byte) io.WriteCloser {
			// log.Println(connID)
			filename := fmt.Sprintf("QUIC_%x.qlog", connID)
			f, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Creating qlog file %s.\n", filename)
			return f
		}
	}
	if common.FatalError(err) != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)

	if common.FatalError(err) != nil {
		return err
	}

	err = common.SetUdpTos(conn, dscp)
	if common.FatalError(err) != nil {
		return err
	}

	tlsConfig, err := common.ServerTLSConfig("/etc", "flowsim-quic")

	if common.FatalError(err) != nil {
		return err
	}

	listener, err := quic.Listen(conn, tlsConfig, quicConf)

	defer listener.Close()
	if common.FatalError(err) != nil {
		return err
	}

	for {

		sess, err := listener.Accept(context.Background())

		if common.FatalError(err) != nil {
			return err
		}
		if single {
			err = quicHandler(sess)
			time.Sleep(500 * time.Millisecond)
			return err
		}

		go quicHandler(sess)
	}

}

// type traceme struct {
// 	quichand quic.Session
// }

// var _ quic.Session = &traceme{}

func quicHandler(sess quic.Session) error {
	log.Println("Entering quicHandler")
	// var qlogger qlog.Tracer

	serverip := sess.LocalAddr()
	hostip := sess.RemoteAddr()

	log.Println(serverip)
	log.Println(hostip)

	// type vers quic.VersionNumber
	// var srcConnID []byte
	// var destConnID []byte
	// var vers quic.VersionNumber
	// log.Println(vers)

	stream, err := sess.AcceptStream(context.Background())
	if common.FatalError(err) != nil {
		return err
	}
	// defer stream.Close()

	// qlogger.StartedConnection(time.Now(), serverip, hostip, vers, srcConnID, destConnID)

	log.Println("Got a stream")

	msgbuf := make([]byte, 128)
	reader := bufio.NewReaderSize(stream, 128)

	for end := false; !end; {
		log.Println("In server loop")
		n, err := reader.Read(msgbuf)
		common.FatalError(err)
		if err != nil {
			if end == true {
				log.Println("Bye!")
				return nil
			}
			return err
		}

		log.Printf("In server loop: got %d bytes: %s", n, msgbuf)
		wbuf, _end, err := parseCmd(string(msgbuf))
		if common.FatalError(err) != nil {
			return err
		}
		end = _end
		_, err = io.Copy(stream, bytes.NewBuffer(wbuf))

		if common.FatalError(err) == nil {
			log.Println("Sent bytes")
			// stat := sess.ConnectionState
			// log.Println(stat)
			// log.Println(connID)
			// log.Println(trace)
		}
	}

	time.Sleep(1 * time.Second)
	return nil
}

// func logq(qs qlog.Tracer) {
// 	es := qs.Export().Error()
// 	qs.StartedConnection(t time.Time, local, remote net.Addr, version protocol.VersionNumber, srcConnID, destConnID protocol.ConnectionID)

// }

// From flowsim TCP
func matcher(cmd string) (string, string, string, error) {
	expr := regexp.MustCompile(`GET (\d+)/(\d+) (\d+)`)
	parsed := expr.FindStringSubmatch(cmd)
	if len(parsed) == 4 {
		return parsed[1], parsed[2], parsed[3], nil
	}
	return "", "", "", errors.New(fmt.Sprintf("Unexpected request %s", cmd))
}

/*
* Purpuse: parse get Command from client
*         and generate a buffer with random bytes
* Return: byte buffer to send or nil on error
*         boolean: true id last bunch
*         error or nil if all went well
*
* Uses crypto/rand, which is already imported for key handling
 */

func parseCmd(strb string) ([]byte, bool, error) {
	// log.Printf("Server: Got %s", strb)

	iter, total, bunchStr, err := matcher(strb)
	if err == nil {
		bunch, _ := strconv.Atoi(bunchStr) // ignore error, wouldn't have parsed the command
		nb := common.RandBytes(bunch)
		if err != nil {
			log.Printf("ERROR while filling random buffer: %v\n", err)
			return nil, iter == total, err
		}
		log.Printf("Sending %d bytes\n", len(nb))
		// exportTraces()
		return nb, iter == total, err
	}
	return nil, false, err
}
