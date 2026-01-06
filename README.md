# ezconf
Library and go generate tool for populating service configurations

## How to use the file loaders

<details>
  <summary>Loading and reading from files</summary>

```golang
	// However we got it, we either have or do not have a path. For our example, let's assume we loaded this from a
	// flag so we end up with a *string which could be nil

	f := file.NoFile()
	if path != nil {
		f = file.SomeFile(*path)
	}

	// Just read the contents of a file. Acts like os.ReadFile(path), but returns an optional Str
	// containing the contents as a string if the file had any contents, // or a None if it was empty.
	// If there was some kind of error (e.g. the file does not exist or is not readable), then the
  // second ok value will be false. This is set up this was so that you can just call
  // contents := f.ReadFile() and use the value without really thinking about it.
	contents, ok := f.ReadFile()
	if !ok {
		fmt.Println("Failed to read from file")
	} else {
		fmt.Println("Got file contents: ", contents)
	}

	// Open a file for reading. File.Open() works just like os.Open(path),
	// so the file is opend in ReadOnly mode.
	var opened *os.File
	opened, err = f.Open()
	if err != nil {
		fmt.Println("Failed to open file for reading: ", err)
		return err
	}
	defer opened.Close()

	// Now use the file handle exactly as you would if you called os.Open()
	return nil
```

</details>

<details>
  <summary>Loading secrets from files</summary>

```golang
	// There is a SecretFile type for convenience since this is a common thing to do
	// in an application. SecretFile simple overrides a few methods of File so that
	// we get a Secret option out of loading the contents instead of a Str.
	f := file.NoSecretFile()
	if path != nil {
		f = file.SomeSecretFile(*path)
	}

	// You can also upgrade a File to a SecretFile
	var normf file.File
	if path != nil {
		normf = file.SomeFile(*path)
	}
	secretf := file.MakeSecret(&normf)

	// You can still see the filepath and everything for a secret file, but we
	// do assume some things about secret files such as the premissions allowed.
	valid, err := secretf.FilePermsValid()
	if err != nil {
		fmt.Println("Failed when validating file permissions for secretf!")
		return err
	}

	if !valid {
		fmt.Println("File permissions for a SecretFile were not 0600!")
	}

	// normf will be cleared as part of upgrading a File to SecretFile
	if normf.IsNone() {
		fmt.Println("Sucessfully cleared normal file after upgrading it to a secret file.")
	}

	// Calling ReadFile() on a SecretFile produces a Secret
	secret, ok := f.ReadFile()
	if !ok {
		fmt.Println("Failed to read from file")
	}

	// This is a secret, so we will only see a redacted value when we try to write it
	// to the console. The same will happen if we try to log it.
	fmt.Println("Got secret file contents: ", secret)

	// Similarly if we try to use stdlib logging libraries
	slog.Info("Second try printing secret file contents", "secret", secret)
	log.Printf("Third try printing secret file contents: %s", secret)
```

</details>

<details>
  <summary>Writing and deleting files</summary>

```golang
	f := file.NoFile()
	if path != nil {
		f = file.SomeFile(*path)
	}

	// Delete a file. Works like os.Remove, but also returns an error if the path is still None
	err := f.Remove()
	if err != nil {
		fmt.Println("Failed to remove file: ", err)
	}

	// Write the contents of a file. Acts like os.WriteFile(path)
	data := []byte("Hello, World!")
	err = f.WriteFile(data, 0644)
	if err != nil {
		fmt.Println("Failed to write file: ", err)
	}

	// Open a file for read/write. File.Create() works like like os.Create(path), which means
	// calling this will either create a file or truncate an existing file. If you want to
	// append to a file, you must use File.OpenFile(os.O_RDWR|os.O_CREATE, 0644) in the same way
	// that would need to when calling os.OpenFile. See https://pkg.go.dev/os#OpenFile for details.
	var opened *os.File
	opened, err = f.Create()
	if err != nil {
		fmt.Println("Failed to open/create file: ", err)
		return err
	}
	defer opened.Close()

	// Now use the file handle exactly as you would if you called os.Create(path)
	opened.Write(data)

	return nil
```

</details>

<details>
  <summary>Other file tools</summary>

```golang
	f := file.NoFile()
	if path != nil {
		f = file.SomeFile(*path)
	}

	// Read back the path
	p, ok := f.Get()
	if ok {
		fmt.Println("Got path: ", p)
	} else {
		fmt.Println("No path given!")
		os.Exit(1)
	}

	// Check if the given path is the same as some other path, matching all equivalent absolute and relative paths.
	// In this case, check if the given path is equivalent to our working directory.
	if f.Match(".") {
		fmt.Println("We are operating on our working directory. Be careful!")
	} else {
		fmt.Println("We are not in our working directory. Go nuts!")
	}

	// Get a new optional with any relative path converted to absolute path (also ensuring it is a valid path)
	abs, err := f.Abs()
	if err != nil {
		fmt.Println("Could not convert path into absolute path. Is it a valid path?")
		return err
	}

	// Stat the file, or just check if it exists if you don't care about other file info
	if abs.Exists() {
		fmt.Println("The file exists!")
	}

	info, err := abs.Stat() // I don't care about the info
	if err != nil {
		fmt.Println("Could not stat the file")
	} else {
		fmt.Println("Got file info: ", info)
	}

	// Check that the file has permissions of at least 0444 (read), but is not 0111 (execute).
	// If those conditions are not fulfilled, we will set perms to 0644.
	valid, err := abs.FilePermsValid(0444, 0111)
	if err != nil {
		fmt.Println("Could not read file permissions!")
		return err
	}

	if !valid {
		err = abs.SetFilePerms(0644)
		if err != nil {
			fmt.Println("Failed to set file perms to 700")
			return err
		}
	}
```

</details>

## How to load certificates and keys

<details>
  <summary>Loading a TLS certificate</summary>

```golang
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
```

</details>

<details>
  <summary>Loading Private and Public keys</summary>

```golang
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
```

</details>


## Secrets

Loading content from a SecretFile will return
an optional.Secret instead of an optional.Str to make things easier for you.

Essentially, this acts as an optional strings for marshalling or otherwise using
the data, but if you try to log or use any print functions on it you will get
a redacted string instead.

## Generating the keys and certs for testing

This is mostly a reminder for myself, given that the certs only have a lifetime of one year.

### RSA

```bash
openssl genrsa -out tls/rsa/key.pem 4096
openssl rsa -in tls/rsa/key.pem -pubout -out tls/rsa/pubkey.pem
openssl req -new -key tls/rsa/key.pem -x509 -sha256 -nodes -subj "/C=US/ST=California/L=Who knows/O=BS Workshops/OU=optional/CN=www.whobe.us" -days 365 -out tls/rsa/cert.pem
```

### ECDSA

```bash
openssl ecparam -name secp521r1 -genkey -noout -out tls/ecdsa/key.pem
openssl ec -in tls/ecdsa/key.pem -pubout > tls/ecdsa/pub.pem
openssl req -new -key tls/ecdsa/key.pem -x509 -sha512 -nodes -subj "/C=US/ST=California/L=Who knows/O=BS Workshops/OU=optional/CN=www.whobe.us" -days 365 -out tls/ecdsa/cert.pem
```

### ED25519

```bash
openssl genpkey -algorithm ed25519 -out tls/ed25519/key.pem
openssl pkey -in tls/ed25519/key.pem -pubout -out tls/ed25519/pub.pem
openssl req -new -key tls/ed25519/key.pem -x509 -nodes -subj "/C=US/ST=California/L=Who knows/O=BS Workshops/OU=optional/CN=www.whobe.us" -days 365 -out tls/ed25519/cert.pem
```
