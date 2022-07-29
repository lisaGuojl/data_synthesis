package hverify

import (
	"crypto/x509"
	"io/ioutil"
	"sync"
	"unsafe"
)

var (
	once        sync.Once
	systemRoots *SMCertPool
)

func systemRootsPool() *SMCertPool {
	once.Do(initSystemRoots)
	return systemRoots
}

var certDirectories = []string{
	"/etc/ssl/certs",               // SLES10/SLES11, https://golang.org/issue/12139
	"/system/etc/security/cacerts", // Android
}
var certFiles = []string{
	"/etc/ssl/certs/ca-certificates.crt", // Debian/Ubuntu/Gentoo etc.
	"/etc/pki/tls/certs/ca-bundle.crt",   // Fedora/RHEL
	"/etc/ssl/ca-bundle.pem",             // OpenSUSE
	"/etc/pki/tls/cacert.pem",            // OpenELEC
}

func initSystemRoots() {
	roots := x509.NewCertPool()
	for _, file := range certFiles {
		data, err := ioutil.ReadFile(file)
		if err == nil {
			roots.AppendCertsFromPEM(data)
			systemRoots = (*SMCertPool)(unsafe.Pointer(roots))
			return
		}
	}

	for _, directory := range certDirectories {
		fis, err := ioutil.ReadDir(directory)
		if err != nil {
			continue
		}
		rootsAdded := false
		for _, fi := range fis {
			data, err := ioutil.ReadFile(directory + "/" + fi.Name())
			if err == nil && roots.AppendCertsFromPEM(data) {
				rootsAdded = true
			}
		}
		if rootsAdded {
			systemRoots = (*SMCertPool)(unsafe.Pointer(roots))
			return
		}
	}
}
