package ezconf

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/brnsampson/optional"
)

// Verifying and setting file permissions for public/private keys and certificates use the following file mode masks.
// The *Perms modes are the desired permissions, while the *PermsMask consts are such that perms && mask should always
// be 0. The mask is only needed because _technically_ I suppose you could make a public key mode 600 or something if you
// really wanted.
const (
	EmptyFilePerms      fs.FileMode = 0000
	CertFilePerms       fs.FileMode = 0600
	CertFilePermsMask   fs.FileMode = 0133
	KeyFilePerms        fs.FileMode = 0600
	PubKeyFilePerms     fs.FileMode = 0644
	KeyFilePermsMask    fs.FileMode = 0177
	PubKeyFilePermsMask fs.FileMode = 0133
)

type pemFile struct {
	File
	setPerms     fs.FileMode
	notPermsMask fs.FileMode
}

func somePem(path string, setPerms, permsNotAllowed fs.FileMode) (pemFile, error) {
	f := SomeFile(path)
	abs, err := f.Abs()
	if err != nil {
		return pemFile{}, err
	}
	return pemFile{abs, setPerms, permsNotAllowed}, nil
}

func noPem(setPerms, permsNotAllowed fs.FileMode) pemFile {
	return pemFile{NoFile(), setPerms, permsNotAllowed}
}

func (o *pemFile) Set(str string) error {
	return o.UnmarshalText([]byte(str))
}

// Override the inner Replace() method to convert path to absolute paths if possible
func (o *pemFile) Replace(path string) optional.Optional[string] {
	abs, err := filepath.Abs(path)
	if err != nil {
		return o.File.Replace(path)
	}

	return o.File.Replace(abs)
}

func (o *pemFile) UnmarshalText(text []byte) error {
	tmp := string(text)
	if tmp == "None" || tmp == "none" || tmp == "null" || tmp == "nil" {
		return o.File.UnmarshalText(text)
	} else {
		_ = o.Replace(tmp)
	}

	return nil
}

func (o pemFile) FilePermsValid() (bool, error) {
	return o.File.FilePermsValid(o.setPerms, o.notPermsMask)
}

func (o pemFile) SetFilePerms() error {
	return o.File.SetFilePerms(o.setPerms)
}

func (o pemFile) ReadBlocks() (blocks []*pem.Block, err error) {
	valid, err := o.FilePermsValid()
	if err != nil {
		return
	}
	if !valid {
		tmp, ok := o.Get()
		if !ok {
			return blocks, fileOptionError("ReadBlocks failed: Path was not set.")
		}
		return blocks, fmt.Errorf("ReadBlocks failed for file %s: Expected file permissions %o", tmp, o.setPerms)
	}

	reader, err := o.Open()
	if err != nil {
		return
	}
	defer reader.Close()

	encoded, err := io.ReadAll(reader)
	if err != nil {
		return
	}

	var block *pem.Block
	for {
		block, encoded = pem.Decode(encoded)
		if block == nil {
			break
		}

		blocks = append(blocks, block)
	}
	return
}

func (o pemFile) WriteBlocks(blocks []*pem.Block) error {
	valid, err := o.FilePermsValid()
	if err != nil {
		return err
	}
	if !valid {
		err := o.SetFilePerms()
		if err != nil {
			return err
		}
	}

	writer, err := o.Create()
	if err != nil {
		return err
	}
	defer writer.Close()

	for _, block := range blocks {
		err = pem.Encode(writer, block)
		if err != nil {
			return err
		}
	}

	return nil
}

// Cert wraps an optional path string and provides extra methods for reading, decoding, and writing pem files containing
// CERTIFICATE blocks.
type Cert struct {
	pemFile
}

func SomeCert(path string) (Cert, error) {
	p, err := somePem(path, CertFilePerms, CertFilePermsMask)
	if err != nil {
		return Cert{}, err
	}
	return Cert{p}, nil
}

func NoCert() Cert {
	return Cert{noPem(CertFilePerms, CertFilePermsMask)}
}

func (o Cert) Type() string {
	return "Cert"
}

func (o *Cert) Set(str string) error {
	return o.UnmarshalText([]byte(str))
}

func (o Cert) String() string {
	if o.IsNone() {
		return "None[Cert]"
	} else {
		tmp, ok := o.Get()
		if !ok {
			return "Error[Cert]"
		}
		return tmp
	}
}

func (o Cert) ReadCerts() (certs []*x509.Certificate, err error) {
	blocks, err := o.ReadBlocks()

	var c []*x509.Certificate
	for _, block := range blocks {
		if block.Type == "CERTIFICATE" {
			c, err = x509.ParseCertificates(block.Bytes)
			if err != nil {
				return
			}
			certs = append(certs, c...)
		}
	}

	return
}

func (o Cert) WriteCerts(certs []*x509.Certificate) error {
	blocks := make([]*pem.Block, 0)
	for _, cert := range certs {
		blocks = append(blocks, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	}

	err := o.WriteBlocks(blocks)
	if err != nil {
		return err
	}
	return nil
}

// PubKey wraps an optional path string and provides extra methods for reading, decoding, and writing pem files containing
// "* PUBLIC KEY" blocks.
type PubKey struct {
	pemFile
}

func SomePubKey(path string) (PubKey, error) {
	p, err := somePem(path, PubKeyFilePerms, PubKeyFilePermsMask)
	if err != nil {
		return PubKey{}, err
	}
	return PubKey{p}, nil
}

func NoPubKey() PubKey {
	return PubKey{noPem(PubKeyFilePerms, PubKeyFilePermsMask)}
}

func (o PubKey) Type() string {
	return "PubKey"
}

func (o *PubKey) Set(str string) error {
	return o.UnmarshalText([]byte(str))
}

func (o PubKey) String() string {
	if o.IsNone() {
		return "None[PubKey]"
	} else {
		tmp, ok := o.Get()
		if !ok {
			return "Error[PubKey]"
		}
		return tmp
	}
}

// ReadPublicKeys will return all public keys found in the given filepath or error. The keys may be of type *rsa.PublicKey,
// *ecdsa.PublicKey, ed25519.PublicKey (Note: that is not a pointer), or *ecdh.PublicKey, depending on the contents of
// the file.
func (o PubKey) ReadPublicKeys() (pub []any, err error) {
	blocks, err := o.ReadBlocks()
	if err != nil {
		return
	}

	var pubKey any
	for _, block := range blocks {
		switch block.Type {
		case "PUBLIC KEY":
			pubKey, err = x509.ParsePKIXPublicKey(block.Bytes)
		case "RSA PUBLIC KEY":
			pubKey, err = x509.ParsePKCS1PublicKey(block.Bytes)
		}
		if err == nil {
			pub = append(pub, pubKey)
		}
	}

	if len(pub) > 0 {
		// There were some public keys in the pem file along with *something* else
		err = nil
	}
	return
}

// WritePublicKey will accept any of an *rsa.PublicKey, *dsa.PublicKey, *ecdsa.PublicKey, ed25519.PublicKey (Note:
// a pointer), or *ecdh.PublicKey. The key will be encoded and written to the path the PubKey option is set to
// with file permissions set appropriately.
func (o PubKey) WritePublicKeys(pubs []any) error {
	blocks := make([]*pem.Block, 0)
	for _, pub := range pubs {
		der, err := x509.MarshalPKIXPublicKey(pub)
		if err != nil {
			return err
		}

		block := pem.Block{Type: "PUBLIC KEY", Bytes: der}
		blocks = append(blocks, &block)
	}

	err := o.WriteBlocks(blocks)
	if err != nil {
		return err
	}
	return nil
}

// PubKey wraps an optional path string and provides extra methods for reading, decoding, and writing pem files containing
// "* PRIVATE KEY" blocks.
type PrivateKey struct {
	pemFile
}

func SomePrivateKey(path string) (PrivateKey, error) {
	p, err := somePem(path, KeyFilePerms, KeyFilePermsMask)
	if err != nil {
		return PrivateKey{}, err
	}
	return PrivateKey{p}, nil
}

func NoPrivateKey() PrivateKey {
	return PrivateKey{noPem(KeyFilePerms, KeyFilePermsMask)}
}

func (o PrivateKey) Type() string {
	return "PrivateKey"
}

func (o *PrivateKey) Set(str string) error {
	return o.UnmarshalText([]byte(str))
}

func (o PrivateKey) String() string {
	if o.IsNone() {
		return "None[PrivateKey]"
	} else {
		tmp, ok := o.Get()
		if !ok {
			return "Error[PrivateKey]"
		}
		return tmp
	}
}

// ReadPrivateKey will return the first private key found in the given filepath or error. This may return an *rsa.PrivateKey,
// *ecdsa.PrivateKey, ed25519.PrivateKey (Note: that is not a pointer), or *ecdh.PrivateKey, depending on the contents of
// the file.
func (o PrivateKey) ReadPrivateKey() (key any, err error) {
	blocks, err := o.ReadBlocks()
	if err != nil {
		return
	}

	var tmp any
	for _, block := range blocks {
		switch block.Type {
		case "PRIVATE KEY":
			tmp, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		case "RSA PRIVATE KEY":
			tmp, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		case "EC PRIVATE KEY":
			tmp, err = x509.ParseECPrivateKey(block.Bytes)
		}
		if err == nil {
			key = tmp
			break
		}
	}
	// if we only encountered errors, err will not be nil. If one of the blocks worked, then error will be nil for that
	// loop and we would have broken at that point.
	return
}

// ReadCert accepts a Cert struct and returns a tls.Certificate for the keypair if both Optionals are Some. This
// is going to be the most used case for anyone loading
func (o PrivateKey) ReadCert(in Cert) (cert tls.Certificate, err error) {
	valid, err := o.FilePermsValid()
	if err != nil {
		return
	}

	keyFile, ok := o.Get()
	if !ok {
		return cert, fileOptionError("ReadCert failed: Keyfile path was not set.")
	}

	if !valid {
		return cert, fmt.Errorf("PrivateKey.ReadCert failed for file %s: Expected file permissions %o", keyFile, o.pemFile.setPerms)
	}

	certFile, ok := in.Get()
	if !ok {
		return cert, fileOptionError("ReadBlocks failed: Certificate path was not set.")
	}

	return tls.LoadX509KeyPair(certFile, keyFile)
}

// WritePrivateKey will accept any of an *rsa.PrivateKey, *dsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey (Note:
// a pointer), or *ecdh.PrivateKey. The key will be encoded and written to the path the PrivateKey option is set to
// with file permissions set appropriately.
func (o PrivateKey) WritePrivateKey(key any) error {
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return err
	}

	block := pem.Block{Type: "PRIVATE KEY", Bytes: der}
	blocks := []*pem.Block{&block}

	err = o.WriteBlocks(blocks)
	if err != nil {
		return err
	}
	return nil
}
