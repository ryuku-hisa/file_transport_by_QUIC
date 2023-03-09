package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/quic-go/quic-go"
)

const addr = "localhost:4242"
const buffSize = 1024
const fname = "./download/data.txt"

// We start a server echoing data on the first stream the client opens,
// then connect with a client, send the message, and wait for its receipt.
func main() {
	fmt.Println("start server...")
	err := server()
	if err != nil {
		log.Fatal(err)
	}
}

// Start a server that echos all data on the first stream opened by the client
func server() error {
	listener, err := quic.ListenAddr(addr, generateTLSConfig(), nil)
	if err != nil {
		return err
	}

	fp, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer fp.Close()

	bw := bufio.NewWriter(fp)

	for {
		fmt.Println("listener accept")
		conn, err := listener.Accept(context.Background())
		if err != nil {
			return err
		}
		fmt.Printf("done\n\n")

		fmt.Println("accept stream")
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			panic(err)
		}
		fmt.Printf("done\n\n")

		buff := make([]byte, buffSize)

		fmt.Println("stream read")
		n, err := stream.Read(buff)
		fmt.Printf("n is %d\n", n)
		if n == 0 {
			fmt.Println("fin reading")
			stream.Close()
			break
		}
		if err != nil {
			return err
		}
		fmt.Printf("done\n\n")

		fmt.Println("write to bufio writer")
		fmt.Println("buff: " + string(buff[:n]))
		if _, err := bw.Write(buff[:n]); err != nil {
			return err
		}
		fmt.Printf("done\n\n")
		stream.Close()
	}

	fmt.Println("flush to file")
	if err = bw.Flush(); err != nil {
		return err
	}
	fmt.Printf("done\n\n")

	fmt.Printf("== DONE server() ==\n\n")
	return nil
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, buffSize)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-file-stream"},
	}
}
