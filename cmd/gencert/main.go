package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"
)

func main() {
	certPath := flag.String("cert", "cert.pem", "path to save the certificate file")
	keyPath := flag.String("key", "key.pem", "path to save the private key file")
	host := flag.String("host", "localhost", "hostname for the certificate")
	validFor := flag.Duration("valid-for", 365*24*time.Hour, "certificate validity duration")
	organization := flag.String("org", "Shortener Dev", "organization name for the certificate")

	flag.Parse()

	if err := generateCertificate(*certPath, *keyPath, *host, *validFor, *organization); err != nil {
		log.Fatalf("Failed to generate certificate: %v", err)
	}

	fmt.Printf("Successfully generated SSL certificate:\n")
	fmt.Printf("  Certificate: %s\n", *certPath)
	fmt.Printf("  Private key: %s\n", *keyPath)
	fmt.Printf("  Host: %s\n", *host)
	fmt.Printf("  Valid for: %v\n", *validFor)
	fmt.Printf("\nYou can now use these certificates with the shortener:\n")
	fmt.Printf("  ./shortener -s -cert-path=%s -key-path=%s\n", *certPath, *keyPath)
}

func generateCertificate(certPath, keyPath, host string, validFor time.Duration, organization string) error {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{organization},
			CommonName:   host,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(validFor),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{host},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	certOut, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	keyOut, err := os.Create(keyPath)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyOut.Close()

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	return nil
}
