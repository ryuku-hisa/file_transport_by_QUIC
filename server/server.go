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
const fname = "./download/data.MOV"

func main() {
	fmt.Println("start server...")
	err := server()
	if err != nil {
		log.Fatal(err)
	}
}

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

	conn, err := listener.Accept(context.Background())
	if err != nil {
		return err
	}

	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	buff := make([]byte, buffSize)

	for {
		n, err := stream.Read(buff)
		if err == io.EOF || n == 0 {
			break
		}
		if err != nil {
			return err
		}

		if _, err := bw.Write(buff[:n]); err != nil {
			return err
		}
	}

	if err = bw.Flush(); err != nil {
		return err
	}
	log.Println("DONE")
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
