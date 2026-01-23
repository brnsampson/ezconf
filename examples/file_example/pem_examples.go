package main

import (
	"context"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/brnsampson/ezconf/file"
)

func TLSCertificate(certPath, keyPath *string) error {
	// Similarly to the File type, the Cert and PrivateKey types make loading and using optional certificates
	// easier and more intuitive. They both embed the Pem struct, which handles the loading of Pem format files.

	// Create a Cert from a flag which requested the user to give the path to the certificate file.
	// Certs and Key Options also return an error if the path cannot be resolved to an
	// absolute path or the file permissions are not correct for a certificate or key file.
	certFile := file.NoCert()
	var err error
	if certPath != nil {
		certFile, err = file.SomeCert(*certPath)
		if err != nil {
			fmt.Println("Failed to initialize cert Option: ", err)
			return err
		}
	}

	// We can use all the same methods as the File type above, but it isn't necessary to go through all of the
	// steps individually. The Cert type knows to check that the path is set, the file exists, and that the file permissions
	// are correct as part of loading the certificates.
	//
	// certificates are returned as a []*x509.Certificate from the file now.
	// Incidentally, we could write new certs to the file with certfile.WriteCerts(certs)
	certs, err := certFile.ReadCerts()
	if err != nil {
		fmt.Println("Error while reading certificates from file: ", err)
		return err
	} else {
		fmt.Println("Found this many certs: ", len(certs))
	}

	// Now we want to load a tls certificate. We typically need two files for this, the certificate(s) and private keyfile.
	// Note: this specifically is for PEM format keys. There are other ways to store keys, but we have not yet implemented
	// support for those. We do support most types of PEM encoded keyfiles though.

	// Certs and Key Options also return an error if the path cannot be resolved to an
	// absolute path or the file permissions are not correct for a certificate or key file.
	var keyFile file.PrivateKey // Effectively the same as privKeyFile := file.NoPrivateKey()
	if keyPath != nil {
		keyFile, err = file.SomePrivateKey(*keyPath)
		if err != nil {
			fmt.Println("Failed to initialize private key Option: ", err)
			return err
		}
	}

	// Again, we could manually do all the validity checks but those are also run as part of loading the TLS certificate.
	// cert is of the type *tls.Certificate, not to be confused with *x509Certificate.
	cert, err := keyFile.ReadCert(certFile)
	if err != nil {
		fmt.Println("Error while generating TLS certificate from PEM format key/cert files: ", err)
		return err
	}

	fmt.Println("Full *tls.Certificate loaded")

	// Now we are ready to start up an TLS sever
	tlsConf := &tls.Config{
		Certificates:             []tls.Certificate{cert},
		MinVersion:               tls.VersionTLS13,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	httpServ := &http.Server{
		Addr:      "127.0.0.1:3000",
		TLSConfig: tlsConf,
	}

	// The parameters ListenAndServeTLS takes are the cert file and keyfile, which may lead you to ask, "why did we bother
	// with all of this then?" Essentially, we were able to do all of our validation and logic with our configuration
	// loading and can put our http server somewhere that makes more sense without just getting panics in our server code
	// when the user passes us an invalid path or something. We are also able to get more granular error messages than just
	// "the server is panicing for some reason."

	fmt.Println("Deferring https server halting for 1 second...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		<-ctx.Done()
		haltctx, haltcancel := context.WithTimeout(context.Background(), time.Second)
		defer haltcancel()
		if err := httpServ.Shutdown(haltctx); err != nil {
			fmt.Println("Error haling http server: ", err)
		}
	}()

	fmt.Println("Starting to listen on https...")
	if err = httpServ.ListenAndServeTLS("", ""); err != nil {
		// This kind of happens even when things go to plan sometimes, so we don't return an error here.
		fmt.Println("TLS server exited with error: ", err)
	}

	return nil
}

func SigningKeys(pubPath, privPath *string) error {
	// In some situations you want to use a public/private keypair for signing instead.
	// Here is how we would load those:
	var privFile file.PrivateKey // Effectively the same as privKeyFile := file.NoPrivateKey()
	var err error
	if privPath != nil {
		privFile, err = file.SomePrivateKey(*privPath)
		if err != nil {
			fmt.Println("Failed to initialize private key Option: ", err)
			return err
		}
	}

	var pubFile file.PubKey // Effectively the same as pubKeyFile := file.NoPubKey()
	if pubPath != nil {
		pubFile, err = file.SomePubKey(*pubPath)
		if err != nil {
			fmt.Println("Failed to initialize private key Option: ", err)
			return err
		}
	}

	// NOTE: as is usually the case with golang key loading, this returns pubKey as a []any and you have to kind of
	// just know how to handle it yourself.
	pubKeys, err := pubFile.ReadPublicKeys()
	if err != nil {
		fmt.Println("Error while reading public key(s) from file: ", err)
		return err
	} else {
		fmt.Println("Found this many public keys: ", len(pubKeys))
	}

	// While a public key file may have multiple public keys, private key files should only have a single key. This
	// key is also returned as an any type which you will then need to sort out how to use just like any other key
	// loading.
	privKey, err := privFile.ReadPrivateKey()
	if err != nil {
		fmt.Println("Error while reading private key from file: ", err)
		return err
	}

	fmt.Println("Loaded a private key from file")
	switch key := privKey.(type) {
	case *rsa.PrivateKey:
		fmt.Println("key is of type RSA:", key)
	case *dsa.PrivateKey:
		fmt.Println("key is of type DSA:", key)
	case *ecdsa.PrivateKey:
		fmt.Println("key is of type ECDSA:", key)
	case ed25519.PrivateKey:
		fmt.Println("key is of type Ed25519:", key)
	default:
		return errors.New("unknown type of private key")
	}

	return nil
}
