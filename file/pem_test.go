package file_test

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/brnsampson/ezconf/file"
	"gotest.tools/v3/assert"
)

func TestCertType(t *testing.T) {
	o, err := file.SomeCert("/not/a/real/path")
	assert.NilError(t, err)

	assert.Equal(t, reflect.TypeOf(o).Name(), o.Type())
}

// Note: this test should always be first because git clone does not preserve file permissions. This means
// this test is also doing to work of setting permissions correctly for other tests.
func TestCertSetFilePerms(t *testing.T) {
	f, err := os.CreateTemp("", "cert")
	assert.NilError(t, err)

	path := f.Name()
	defer os.Remove(path)
	f.Chmod(0666)

	o, err := file.SomeCert(path)
	assert.NilError(t, err)
	err = o.SetFilePerms()
	assert.NilError(t, err)
	s, err := f.Stat()
	assert.NilError(t, err)
	assert.Equal(t, file.CertFilePerms, s.Mode())

	// Now that the sanity of the function itself has been checked, let's apply it to the file we are testing against.
	path = "../testing/rsa/cert.pem"
	o, err = file.SomeCert(path)
	assert.NilError(t, err)
	err = o.SetFilePerms()
	assert.NilError(t, err)

	path = "../testing/ecdsa/cert.pem"
	o, err = file.SomeCert(path)
	assert.NilError(t, err)
	err = o.SetFilePerms()
	assert.NilError(t, err)

	path = "../testing/ed25519/cert.pem"
	o, err = file.SomeCert(path)
	assert.NilError(t, err)
	err = o.SetFilePerms()
	assert.NilError(t, err)
}

func TestCertGet(t *testing.T) {
	path := "../testing/rsa/cert.pem"
	o, err := file.SomeCert(path)
	assert.NilError(t, err)
	abs, err := filepath.Abs(path)
	// an error here doesn't mean our library is broken, just that the path we chose to test with isn't valid.
	assert.NilError(t, err)

	tmp, ok := o.Get()
	assert.Assert(t, ok)

	assert.Equal(t, abs, tmp)
}

func TestCertString(t *testing.T) {
	path := "../testing/rsa/cert.pem"
	abs, err := filepath.Abs(path)
	// an error here doesn't mean our library is broken, just that the path we chose to test with isn't valid.
	assert.NilError(t, err)

	o, err := file.SomeCert(path)
	assert.NilError(t, err)
	assert.Equal(t, abs, o.String())
}

func TestCertMarshalText(t *testing.T) {
	path := "../testing/rsa/cert.pem"
	abs, err := filepath.Abs(path)
	// an error here doesn't mean our library is broken, just that the path we chose to test with isn't valid.
	assert.NilError(t, err)

	o, err := file.SomeCert(path)
	assert.NilError(t, err)

	s, err := o.MarshalText()
	assert.NilError(t, err)
	assert.Equal(t, abs, string(s))
}

func TestCertUnmarshalText(t *testing.T) {
	path := "../testing/rsa/cert.pem"
	nullCert := "null"
	intCert := "42"
	abs, err := filepath.Abs(path)
	// an error here doesn't mean our library is broken, just that the path we chose to test with isn't valid.
	assert.NilError(t, err)
	intAbs, err := filepath.Abs(intCert)
	// an error here doesn't mean our library is broken, just that the path we chose to test with isn't valid.
	assert.NilError(t, err)

	// Text sucessful unmarshaling
	o := file.NoCert()
	err = o.UnmarshalText([]byte(path))
	assert.NilError(t, err)
	assert.Assert(t, o.IsSome())

	ret, ok := o.Get()
	assert.Assert(t, ok)
	assert.Equal(t, abs, ret)

	// Test unmarshaling null
	err = o.UnmarshalText([]byte(nullCert))
	assert.NilError(t, err)
	assert.Assert(t, o.IsNone())

	// Test unmarshaling non-path. This will work because it should interpret this as a file named "41", which you could
	// theoretically have.
	err = o.UnmarshalText([]byte(intCert))
	assert.NilError(t, err)

	ret, ok = o.Get()
	assert.Assert(t, ok)
	assert.Equal(t, intAbs, ret)
}

func TestCertFilePermsValid(t *testing.T) {
	valid_path := "../testing/rsa/cert.pem"
	invalid_path := "../testing/rsa/cert_bad_perms.pem"

	// Make sure that the invalid file perms are actually set!
	err := os.Chmod(invalid_path, 0666)
	assert.NilError(t, err)

	v, err := file.SomeCert(valid_path)
	assert.NilError(t, err)
	i, err := file.SomeCert(invalid_path)
	assert.NilError(t, err)

	good, err := v.FilePermsValid()
	assert.NilError(t, err)

	bad, err := i.FilePermsValid()
	assert.NilError(t, err)

	assert.Assert(t, good)
	assert.Assert(t, !bad)
}

func TestCertReadRSACerts(t *testing.T) {
	expectedCertIssuer := "CN=www.whobe.us,OU=optional,O=BS Workshops,L=Who knows,ST=California,C=US"
	certPath := "../testing/rsa/cert.pem"

	c, err := file.SomeCert(certPath)
	assert.NilError(t, err)

	certs, err := c.ReadCerts()
	assert.NilError(t, err)
	assert.Assert(t, len(certs) == 1)

	cert := certs[0]
	assert.Equal(t, x509.SHA256WithRSA, cert.SignatureAlgorithm)
	assert.Equal(t, expectedCertIssuer, cert.Issuer.String())
}

func TestCertReadECDSACerts(t *testing.T) {
	expectedCertIssuer := "CN=www.whobe.us,OU=optional,O=BS Workshops,L=Who knows,ST=California,C=US"
	certPath := "../testing/ecdsa/cert.pem"

	c, err := file.SomeCert(certPath)
	assert.NilError(t, err)

	certs, err := c.ReadCerts()
	assert.NilError(t, err)
	assert.Assert(t, len(certs) == 1)

	cert := certs[0]
	assert.Equal(t, x509.ECDSAWithSHA512, cert.SignatureAlgorithm)
	assert.Equal(t, expectedCertIssuer, cert.Issuer.String())
}

func TestCertReadED25519Certs(t *testing.T) {
	expectedCertIssuer := "CN=www.whobe.us,OU=optional,O=BS Workshops,L=Who knows,ST=California,C=US"
	certPath := "../testing/ed25519/cert.pem"

	c, err := file.SomeCert(certPath)
	assert.NilError(t, err)

	certs, err := c.ReadCerts()
	assert.NilError(t, err)
	assert.Assert(t, len(certs) == 1)

	cert := certs[0]
	assert.Equal(t, x509.PureEd25519, cert.SignatureAlgorithm)
	assert.Equal(t, expectedCertIssuer, cert.Issuer.String())
}

func TestCertFileWriteCerts(t *testing.T) {
	// Read in valid certificates (tested above)
	certPath := "../testing/rsa/cert.pem"
	c, err := file.SomeCert(certPath)
	assert.NilError(t, err)

	certs, err := c.ReadCerts()
	assert.NilError(t, err)

	f, err := os.CreateTemp("", "cert")
	assert.NilError(t, err)

	// Create temporary file (empty)
	path := f.Name()
	defer os.Remove(path)
	f.Chmod(0644)
	f.Close()

	tc, err := file.SomeCert(path)
	assert.NilError(t, err)

	// Write certificates to new file
	err = tc.WriteCerts(certs)
	assert.NilError(t, err)

	// Check that reading the certs back gives us the expected values
	newCerts, err := tc.ReadCerts()
	assert.NilError(t, err)

	// validate that we are testing what we think we are testing
	assert.Assert(t, len(certs) == 1)
	assert.Assert(t, len(newCerts) == 1)
	assert.Assert(t, certs[0].Equal(newCerts[0]))
}

func TestPubKeyType(t *testing.T) {
	o, err := file.SomePubKey("/not/a/real/path")
	assert.NilError(t, err)
	assert.Equal(t, reflect.TypeOf(o).Name(), o.Type())
}

func TestPubKeySetFilePerms(t *testing.T) {
	f, err := os.CreateTemp("", "pubkey")
	assert.NilError(t, err)

	path := f.Name()
	defer os.Remove(path)
	f.Chmod(0666)

	o, err := file.SomePubKey(path)
	assert.NilError(t, err)
	o.SetFilePerms()
	s, err := f.Stat()
	assert.NilError(t, err)
	assert.Equal(t, file.PubKeyFilePerms, s.Mode())

	// Now that the sanity of the function itself has been checked, let's apply it to the file we are testing against.
	path = "../testing/rsa/pubkey.pem"
	o, err = file.SomePubKey(path)
	assert.NilError(t, err)
	err = o.SetFilePerms()
	assert.NilError(t, err)

	path = "../testing/ecdsa/pub.pem"
	o, err = file.SomePubKey(path)
	assert.NilError(t, err)
	err = o.SetFilePerms()
	assert.NilError(t, err)

	path = "../testing/ed25519/pub.pem"
	o, err = file.SomePubKey(path)
	assert.NilError(t, err)
	err = o.SetFilePerms()
	assert.NilError(t, err)
}

func TestPubKeyString(t *testing.T) {
	path := "../testing/rsa/pubkey.pem"
	abs, err := filepath.Abs(path)
	// an error here doesn't mean our library is broken, just that the path we chose to test with isn't valid.
	assert.NilError(t, err)

	o, err := file.SomePubKey(path)
	assert.NilError(t, err)
	assert.Equal(t, abs, o.String())
}

func TestPubKeyMarshalText(t *testing.T) {
	path := "../testing/rsa/pubkey.pem"
	abs, err := filepath.Abs(path)
	// an error here doesn't mean our library is broken, just that the path we chose to test with isn't valid.
	assert.NilError(t, err)

	o, err := file.SomePubKey(path)
	assert.NilError(t, err)

	s, err := o.MarshalText()
	assert.NilError(t, err)
	assert.Equal(t, abs, string(s))
}

func TestPubKeyUnmarshalText(t *testing.T) {
	path := "../testing/rsa/pubkey.pem"
	nullPubKey := "null"
	intPubKey := "42"
	abs, err := filepath.Abs(path)
	// an error here doesn't mean our library is broken, just that the path we chose to test with isn't valid.
	assert.NilError(t, err)
	intAbs, err := filepath.Abs(intPubKey)
	// an error here doesn't mean our library is broken, just that the path we chose to test with isn't valid.
	assert.NilError(t, err)

	// Text sucessful unmarshaling
	o := file.NoPubKey()
	err = o.UnmarshalText([]byte(path))
	assert.NilError(t, err)
	assert.Assert(t, o.IsSome())

	ret, ok := o.Get()
	assert.Assert(t, ok)
	assert.Equal(t, abs, ret)

	// Test unmarshaling null
	err = o.UnmarshalText([]byte(nullPubKey))
	assert.NilError(t, err)
	assert.Assert(t, o.IsNone())

	// Test unmarshaling non-path. This will work because it should interpret this as a file named "41", which you could
	// theoretically have.
	err = o.UnmarshalText([]byte(intPubKey))
	assert.NilError(t, err)

	ret, ok = o.Get()
	assert.Assert(t, ok)
	assert.Equal(t, intAbs, ret)
}

func TestPubKeyFilePermsValid(t *testing.T) {
	valid_path := "../testing/rsa/pubkey.pem"
	invalid_path := "../testing/rsa/pubkey_bad_perms.pem"

	// Make sure that the invalid file perms are actually set!
	err := os.Chmod(invalid_path, 0666)
	assert.NilError(t, err)

	v, err := file.SomePubKey(valid_path)
	assert.NilError(t, err)
	i, err := file.SomePubKey(invalid_path)
	assert.NilError(t, err)

	good, err := v.FilePermsValid()
	assert.NilError(t, err)

	bad, err := i.FilePermsValid()
	assert.NilError(t, err)

	assert.Assert(t, good)
	assert.Assert(t, !bad)
}

func TestPubKeyReadPublicKeysRSA(t *testing.T) {
	pubPath := "../testing/rsa/pubkey.pem"
	p, err := file.SomePubKey(pubPath)
	assert.NilError(t, err)
	keys, err := p.ReadPublicKeys()
	assert.NilError(t, err)

	// There is only one cert in the file...
	assert.Assert(t, len(keys) == 1)
	key := keys[0]

	switch key.(type) {
	case *rsa.PublicKey:
	default:
		panic("rsa pem file was not read into rsa key!")
	}
}

func TestPubKeyReadPublicKeysECDSA(t *testing.T) {
	pubPath := "../testing/ecdsa/pub.pem"

	p, err := file.SomePubKey(pubPath)
	assert.NilError(t, err)
	keys, err := p.ReadPublicKeys()
	assert.NilError(t, err)

	// There is only one cert in the file...
	assert.Assert(t, len(keys) == 1)
	key := keys[0]

	switch key.(type) {
	case *ecdsa.PublicKey:
	default:
		panic("rsa pem file was not read into rsa key!")
	}
}

func TestPubKeyReadPublicKeysED25519(t *testing.T) {
	pubPath := "../testing/ed25519/pub.pem"

	p, err := file.SomePubKey(pubPath)
	assert.NilError(t, err)
	keys, err := p.ReadPublicKeys()
	assert.NilError(t, err)

	// There is only one cert in the file...
	assert.Assert(t, len(keys) == 1)
	key := keys[0]

	switch key.(type) {
	case ed25519.PublicKey:
	default:
		panic("rsa pem file was not read into rsa key!")
	}
}

func TestPubKeyWritePubKey(t *testing.T) {
	// Read in valid certificates (tested above)
	pubPath := "../testing/rsa/pubkey.pem"
	c, err := file.SomePubKey(pubPath)
	assert.NilError(t, err)
	keys, err := c.ReadPublicKeys()
	assert.NilError(t, err)

	f, err := os.CreateTemp("", "pubkey")
	assert.NilError(t, err)

	// Create temporary file (empty)
	path := f.Name()
	defer os.Remove(path)
	f.Chmod(0666)
	f.Close()

	tc, err := file.SomePubKey(path)
	assert.NilError(t, err)
	// Write certificates to new file
	err = tc.WritePublicKeys(keys)
	assert.NilError(t, err)

	// Check that reading the certs back gives us the expected values
	newKeys, err := tc.ReadPublicKeys()
	assert.NilError(t, err)

	// validate that we are testing what we think we are testing
	assert.Assert(t, len(keys) == 1)
	assert.Assert(t, len(newKeys) == 1)

	key := keys[0]
	newKey := newKeys[0]
	switch k1 := key.(type) {
	case *rsa.PublicKey:
		switch k2 := newKey.(type) {
		case *rsa.PublicKey:
			assert.Assert(t, k1.Equal(k2))
		default:
			panic("Expected key loaded from tmp file to be *rsa.PublicKey, but it wasn't!")
		}
	default:
		panic("Expected key loaded from tmp file to be *rsa.PublicKey, but it wasn't!")
	}
}

func TestPrivateKeyType(t *testing.T) {
	o, err := file.SomePrivateKey("/not/a/real/path")
	assert.NilError(t, err)
	assert.Equal(t, reflect.TypeOf(o).Name(), o.Type())
}

func TestPrivateKeySetFilePerms(t *testing.T) {
	f, err := os.CreateTemp("", "key")
	assert.NilError(t, err)

	path := f.Name()
	defer os.Remove(path)
	f.Chmod(0666)

	o, err := file.SomePrivateKey(path)
	assert.NilError(t, err)
	o.SetFilePerms()
	s, err := f.Stat()
	assert.NilError(t, err)
	assert.Equal(t, file.KeyFilePerms, s.Mode())

	// Now that the sanity of the function itself has been checked, let's apply it to the file we are testing against.
	path = "../testing/rsa/key.pem"
	o, err = file.SomePrivateKey(path)
	assert.NilError(t, err)
	err = o.SetFilePerms()
	assert.NilError(t, err)

	path = "../testing/ecdsa/key.pem"
	o, err = file.SomePrivateKey(path)
	assert.NilError(t, err)
	err = o.SetFilePerms()
	assert.NilError(t, err)

	path = "../testing/ed25519/key.pem"
	o, err = file.SomePrivateKey(path)
	assert.NilError(t, err)
	err = o.SetFilePerms()
	assert.NilError(t, err)
}

func TestPrivateKeyString(t *testing.T) {
	path := "../testing/rsa/key.pem"
	abs, err := filepath.Abs(path)
	// an error here doesn't mean our library is broken, just that the path we chose to test with isn't valid.
	assert.NilError(t, err)

	o, err := file.SomePrivateKey(path)
	assert.NilError(t, err)
	assert.Equal(t, abs, o.String())
}

func TestPrivateKeyMarshalText(t *testing.T) {
	path := "../testing/rsa/key.pem"
	abs, err := filepath.Abs(path)
	// an error here doesn't mean our library is broken, just that the path we chose to test with isn't valid.
	assert.NilError(t, err)

	o, err := file.SomePrivateKey(path)
	assert.NilError(t, err)

	s, err := o.MarshalText()
	assert.NilError(t, err)
	assert.Equal(t, abs, string(s))
}

func TestPrivateKeyUnmarshalText(t *testing.T) {
	path := "../testing/rsa/key.pem"
	nullPrivateKey := "null"
	intPrivateKey := "42"
	abs, err := filepath.Abs(path)
	// an error here doesn't mean our library is broken, just that the path we chose to test with isn't valid.
	assert.NilError(t, err)
	intAbs, err := filepath.Abs(intPrivateKey)
	// an error here doesn't mean our library is broken, just that the path we chose to test with isn't valid.
	assert.NilError(t, err)

	// Text sucessful unmarshaling
	o := file.NoPrivateKey()
	err = o.UnmarshalText([]byte(path))
	assert.NilError(t, err)
	assert.Assert(t, o.IsSome())

	ret, ok := o.Get()
	assert.Assert(t, ok)
	assert.Equal(t, abs, ret)

	// Test unmarshaling null
	err = o.UnmarshalText([]byte(nullPrivateKey))
	assert.NilError(t, err)
	assert.Assert(t, o.IsNone())

	// Test unmarshaling non-path. This will work because it should interpret this as a file named "41", which you could
	// theoretically have.
	err = o.UnmarshalText([]byte(intPrivateKey))
	assert.NilError(t, err)

	ret, ok = o.Get()
	assert.Assert(t, ok)
	assert.Equal(t, intAbs, ret)
}

func TestPrivateKeyFilePermsValid(t *testing.T) {
	valid_path := "../testing/rsa/key.pem"
	invalid_path := "../testing/rsa/key_bad_perms.pem"

	// Make sure that the invalid file perms are actually set!
	err := os.Chmod(invalid_path, 0666)
	assert.NilError(t, err)

	v, err := file.SomePrivateKey(valid_path)
	assert.NilError(t, err)
	i, err := file.SomePrivateKey(invalid_path)
	assert.NilError(t, err)

	good, err := v.FilePermsValid()
	assert.NilError(t, err)

	bad, err := i.FilePermsValid()
	assert.NilError(t, err)

	assert.Assert(t, good)
	assert.Assert(t, !bad)
}

func TestPrivateKeyReadCert(t *testing.T) {
	keyPath := "../testing/rsa/key.pem"
	certPath := "../testing/rsa/cert.pem"
	ko, err := file.SomePrivateKey(keyPath)
	assert.NilError(t, err)
	co, err := file.SomeCert(certPath)
	assert.NilError(t, err)

	key, err := ko.ReadPrivateKey()
	certificate, err := ko.ReadCert(co)
	assert.NilError(t, err)

	// Check that the cert loaded correctly
	certs, err := co.ReadCerts()
	assert.Assert(t, len(certs) == 1)
	assert.Assert(t, len(certificate.Certificate) == 1)
	assert.Assert(t, bytes.Equal(certs[0].Raw, certificate.Certificate[0]))

	// Check that the private key loaded correctly
	type privKeyInter interface {
		Public() crypto.PublicKey
		Equal(x crypto.PrivateKey) bool
	}

	switch k1 := key.(type) {
	case privKeyInter:
		switch k2 := certificate.PrivateKey.(type) {
		case privKeyInter:
			assert.Assert(t, k1.Equal(k2))
		default:
			panic("Expected *rsa.PrivateKey to implement the crypto.PrivateKey interface, but it didn't!")
		}
	default:
		panic("Expected key from tls.Certificate to implement the crypto.PrivateKey interface, but it didn't!")
	}
}

func TestPrivateKeyReadPrivateKeyRSA(t *testing.T) {
	keyPath := "../testing/rsa/key.pem"
	p, err := file.SomePrivateKey(keyPath)
	assert.NilError(t, err)
	key, err := p.ReadPrivateKey()
	assert.NilError(t, err)

	switch key.(type) {
	case *rsa.PrivateKey:
	default:
		panic("rsa pem file was not read into the correct format key!")
	}
}

func TestPrivateKeyReadPrivateKeyECDSA(t *testing.T) {
	keyPath := "../testing/ecdsa/key.pem"

	p, err := file.SomePrivateKey(keyPath)
	assert.NilError(t, err)
	key, err := p.ReadPrivateKey()
	assert.NilError(t, err)

	switch key.(type) {
	case *ecdsa.PrivateKey:
	default:
		panic("ecdsa pem file was not read into the correct format key!")
	}
}

func TestPrivateKeyReadPrivateKeyED25519(t *testing.T) {
	keyPath := "../testing/ed25519/key.pem"

	p, err := file.SomePrivateKey(keyPath)
	assert.NilError(t, err)
	key, err := p.ReadPrivateKey()
	assert.NilError(t, err)

	switch key.(type) {
	case ed25519.PrivateKey:
	default:
		panic("ed25519 pem file was not read into the correct format key!")
	}
}

func TestPrivateKeyWritePrivateKey(t *testing.T) {
	// Read in valid certificates (tested above)
	keyPath := "../testing/rsa/key.pem"
	c, err := file.SomePrivateKey(keyPath)
	assert.NilError(t, err)
	key, err := c.ReadPrivateKey()
	assert.NilError(t, err)

	f, err := os.CreateTemp("", "key")
	assert.NilError(t, err)

	// Create temporary file (empty)
	path := f.Name()
	defer os.Remove(path)
	f.Chmod(0666)
	f.Close()

	tc, err := file.SomePrivateKey(path)
	assert.NilError(t, err)
	// Write certificates to new file
	err = tc.WritePrivateKey(key)
	assert.NilError(t, err)

	// Check that reading the certs back gives us the expected values
	newKey, err := tc.ReadPrivateKey()
	assert.NilError(t, err)

	switch k1 := key.(type) {
	case *rsa.PrivateKey:
		switch k2 := newKey.(type) {
		case *rsa.PrivateKey:
			assert.Assert(t, k1.Equal(k2))
		default:
			panic("Expected key loaded from tmp file to be *rsa.PrivateKey, but it wasn't!")
		}
	default:
		panic("Expected key loaded from tmp file to be *rsa.PrivateKey, but it wasn't!")
	}
}
