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
	"io"
	"log"
	"math/big"
	"os"

	"github.com/quic-go/quic-go"
)

const addr = "localhost:50051"
const buffSize = 1024
const fname = "./download/data.txt"

func main() {
	fmt.Println("start server...")
	server()
}

func server() error {
	listener, err := quic.ListenAddr(addr, generateTLSConfig(), nil)
	if err != nil {
		log.Fatal(err)
	}
	for {
		fp, err := os.Create(fname)
		if err != nil {
			log.Fatal(err)
		}
		defer fp.Close()

		conn, err := listener.Accept(context.Background())
		log.Println("Listener accept")
		if err != nil {
			log.Fatal(err)
		}
		defer listener.Close()

		bw := bufio.NewWriter(fp)
		go receive_handler(bw, conn)
	}

}

func receive_handler(bw *bufio.Writer, conn quic.Connection) {

	stream, err := conn.AcceptStream(context.Background())
	log.Println("Accept Stream")
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		log.Println("Stream Closed")
		stream.Close()
	}()

	log.Println("Start Copying... ")
	_, err = io.Copy(bw, stream)
	log.Println("Copying DONE")
	if err != nil {
		log.Fatal(err)
	}

}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
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
