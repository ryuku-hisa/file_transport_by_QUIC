package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/quic-go/quic-go"
)

const addr = "localhost:50051"

func main() {
	if len(os.Args) != 2 {
		log.Fatal("invalid argument")
	}
	fname := os.Args[1]
	err := sender(fname)
	if err != nil {
		panic(err)
	}
}

func sender(fname string) error {
	fp, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-file-stream"},
	}

	conn, err := quic.DialAddr(addr, tlsConf, nil)

	// stream, err := conn.OpenStreamSync(context.Background())
	stream, err := conn.OpenStream()
	defer stream.Close()

	for {
		n, err := io.Copy(stream, fp)
		fmt.Printf("n is %d\n", n)
		if err != nil {
			return err
		}
		if n == 0 {
			break
		}
	}
	return nil
}
