package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
  "io"

	"github.com/quic-go/quic-go"
)

const addr = "localhost:4242"

// We start a server echoing data on the first stream the client opens,
// then connect with a client, send the message, and wait for its receipt.
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

	buff := make([]byte, 1024)

	for {
		n, err := fp.Read(buff)
		if n == 0 {
			break
		}
		if err != nil {
			panic(err)
		}
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}
	conn, err := quic.DialAddr(addr, tlsConf, nil)
	if err != nil {
		return err
	}

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("Sending data...")
	_, err = stream.Write(buff)
	if err != nil {
		return err
	}
  _, err = io.ReadFull(stream, buff)
  if err != nil {
    return err
  }

  fmt.Println("DONE")

	return nil
}
