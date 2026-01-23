package main

import (
	"fmt"
	"os"
)

func main() {
	// Basic Options
	code := 0

	// File Options
	path := "example.txt"
	err := LoadingAndReadingFiles(&path)
	if err != nil {
		fmt.Println("LoadingAndReadingFiles example failed")
		code = 1
	}

	err = SecretFiles(&path)
	if err != nil {
		fmt.Println("SecretFiles example failed")
		code = 1
	}

	err = WritingAndDeletingFiles(&path)
	if err != nil {
		fmt.Println("WritingAndDeletingFiles example failed")
		code = 1
	}

	err = AdditionalFileTools(&path)
	if err != nil {
		fmt.Println("AdditionalFileTools example failed")
		code = 1
	}

	// Cert and Key Options
	certPath := "../../testing/rsa/cert.pem"
	keyPath := "../../testing/rsa/key.pem"
	pubKeyPath := "../../testing/ed25519/pub.pem"
	privKeyPath := "../../testing/ed25519/key.pem"
	err = TLSCertificate(&certPath, &keyPath)
	if err != nil {
		fmt.Println("TLSCertificate example failed")
		code = 1
	}

	err = SigningKeys(&pubKeyPath, &privKeyPath)
	if err != nil {
		fmt.Println("SigningKeys example failed")
		code = 1
	}

	if code == 0 {
		fmt.Println("")
		fmt.Println("Examples ran successfully")
	} else {
		fmt.Println("")
		fmt.Println("At least one example failed")
	}
	os.Exit(code)
}
