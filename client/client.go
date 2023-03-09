package main

import (
	"context"
	"crypto/tls"
	"fmt"
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
	err := client(fname)
	if err != nil {
		panic(err)
	}
}

func client(fname string) error {
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

	stream, err := conn.OpenStreamSync(context.Background())
	defer stream.Close()

	buff := make([]byte, 1024)

	for {

		n, err := fp.Read(buff)
		if n == 0 {
			fmt.Println("fin reading")
			break
		}
		if err != nil {
			return err
		}

		_, err = stream.Write(buff[:n])
		if err != nil {
			return err
		}
	}
	log.Println("DONE")

	return nil
}
