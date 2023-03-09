package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"

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

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-file-stream"},
	}

	fmt.Println("dial address")
	conn, err := quic.DialAddr(addr, tlsConf, nil)

	stream, err := conn.OpenStream()
	for {
		fmt.Println("read data from the file")
		buff := make([]byte, 1024)
		n, err := fp.Read(buff)
		fmt.Printf("n is %d\n", n)
		// if n == 0 {
		// 	fmt.Println("fin reading")
		// 	break
		// }
		if err != nil {
			panic(err)
		}
		fmt.Printf("done\n\n")

		fmt.Println("stream write")
		_, err = stream.Write(buff[:n])
		if err != nil {
			return err
		}
		fmt.Printf("done\n\n")

		fmt.Println("send data")
		_, err = io.ReadFull(stream, buff[:n])
		if err == io.EOF {
			fmt.Println("fin reading")
			break
		}
		if err != nil {
			return err
		}
		fmt.Printf("done\n\n")

		fmt.Println("accept stream")
		conn.AcceptStream(context.Background())
		fmt.Printf("done\n\n")
	}

	stream.Close()
	fmt.Println("DONE")

	return nil
}
