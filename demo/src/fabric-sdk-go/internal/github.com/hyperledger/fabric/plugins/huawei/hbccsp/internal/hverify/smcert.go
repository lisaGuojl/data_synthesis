package hverify

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/plugins/huawei/hbccsp/internal/hsm"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/plugins/huawei/hbccsp/internal/sx509"
	"math/big"
	"net"
	"runtime"
	"strings"
	"time"
	"unsafe"

	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
	"github.com/pkg/errors"
)

var logger = flogging.MustGetLogger("hverify")

const (
	leafCertificate = iota
	intermediateCertificate
	rootCertificate
)

type SMVerifyOptions struct {
	DNSName       string
	Intermediates *SMCertPool
	Roots         *SMCertPool // if nil, the system roots are used
	CurrentTime   time.Time   // if zero, the current time is used

	KeyUsages []x509.ExtKeyUsage
}
type SMCertPool struct {
	bySubjectKeyId map[string][]int
	byName         map[string][]int
	certs          []*SMCertificate
}

type SMCertificate struct {
	Raw                     []byte // Complete ASN.1 DER content (certificate, signature algorithm and signature).
	RawTBSCertificate       []byte // Certificate part of raw ASN.1 DER content.
	RawSubjectPublicKeyInfo []byte // DER encoded SubjectPublicKeyInfo.
	RawSubject              []byte // DER encoded Subject
	RawIssuer               []byte // DER encoded Issuer

	Signature          []byte
	SignatureAlgorithm x509.SignatureAlgorithm

	PublicKeyAlgorithm x509.PublicKeyAlgorithm
	PublicKey          interface{}

	Version             int
	SerialNumber        *big.Int
	Issuer              pkix.Name
	Subject             pkix.Name
	NotBefore, NotAfter time.Time // Validity bounds.
	KeyUsage            x509.KeyUsage

	Extensions []pkix.Extension

	ExtraExtensions []pkix.Extension

	UnhandledCriticalExtensions []asn1.ObjectIdentifier

	ExtKeyUsage        []x509.ExtKeyUsage      // Sequence of extended key usages.
	UnknownExtKeyUsage []asn1.ObjectIdentifier // Encountered extended key usages unknown to this package.

	BasicConstraintsValid bool // if true then the next two fields are valid.
	IsCA                  bool
	MaxPathLen            int

	MaxPathLenZero bool

	SubjectKeyId   []byte
	AuthorityKeyId []byte

	// RFC 5280, 4.2.2.1 (Authority Information Access)
	OCSPServer            []string
	IssuingCertificateURL []string

	// Subject Alternate Name values
	DNSNames       []string
	EmailAddresses []string
	IPAddresses    []net.IP

	// Name constraints
	PermittedDNSDomainsCritical bool // if true then the name constraints are marked critical.
	PermittedDNSDomains         []string

	// CRL Distribution Points
	CRLDistributionPoints []string

	PolicyIdentifiers []asn1.ObjectIdentifier
}

type UnknownAuthorityError struct {
	cert *x509.Certificate
	// hintErr contains an error that may be helpful in determining why an
	// authority wasn't found.
	hintErr error
	// hintCert contains a possible authority certificate that was rejected
	// because of the error in hintErr.
	hintCert *x509.Certificate
}

func (e UnknownAuthorityError) Error() string {
	s := "x509: certificate signed by unknown authority"
	if e.hintErr != nil {
		certName := e.hintCert.Subject.CommonName
		if len(certName) == 0 {
			if len(e.hintCert.Subject.Organization) > 0 {
				certName = e.hintCert.Subject.Organization[0]
			}
			certName = "serial:" + e.hintCert.SerialNumber.String()
		}
		s += fmt.Sprintf(" (possibly because of %q while trying to verify candidate authority certificate %q)", e.hintErr, certName)
	}
	return s
}

var errNotParsed = errors.New("x509: missing ASN.1 contents; use ParseCertificate")

func (smc *SMCertificate) SMVerify(optsx x509.VerifyOptions) (chains [][]*x509.Certificate, err error) {
	c := (*x509.Certificate)(unsafe.Pointer(smc))
	opts := *(*SMVerifyOptions)(unsafe.Pointer(&optsx))

	if len(c.Raw) == 0 {
		return nil, errNotParsed
	}
	if opts.Intermediates != nil {
		for _, intermediate := range opts.Intermediates.certs {
			if len(intermediate.Raw) == 0 {
				return nil, errNotParsed
			}
		}
	}

	if opts.Roots == nil && runtime.GOOS == "windows" {
		return nil, nil
	}

	if len(c.UnhandledCriticalExtensions) > 0 {
		return nil, x509.UnhandledCriticalExtension{}
	}

	if opts.Roots == nil {
		opts.Roots = systemRootsPool()
		if opts.Roots == nil {
			return nil, x509.SystemRootsError{}
		}
	}

	err = smc.isValid(leafCertificate, nil, &opts)
	if err != nil {
		return nil, errors.WithMessage(err, "check cert valid")
	}

	if len(opts.DNSName) > 0 {
		err = c.VerifyHostname(opts.DNSName)
		if err != nil {
			return nil, errors.WithMessage(err, "veriy hostname")
		}
	}
	var candidateChains [][]*x509.Certificate
	if opts.Roots.contains(c) {
		candidateChains = append(candidateChains, []*x509.Certificate{c})
	} else {
		candidateChains, err = smc.buildChains(make(map[int][][]*x509.Certificate), []*x509.Certificate{c}, &opts)
		if err != nil {
			return nil, errors.WithMessage(err, "buildchains fail")
		}
	}

	keyUsages := opts.KeyUsages
	if len(keyUsages) == 0 {
		keyUsages = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	}

	for _, usage := range keyUsages {
		if usage == x509.ExtKeyUsageAny {
			chains = candidateChains
			return chains, nil
		}
	}

	for _, candidate := range candidateChains {
		if checkChainForKeyUsage(candidate, keyUsages) {
			chains = append(chains, candidate)
		}
	}

	if len(chains) == 0 {
		return nil, x509.CertificateInvalidError{Cert: c, Reason: x509.IncompatibleUsage}
	}

	return chains, nil
}

func checkChainForKeyUsage(chain []*x509.Certificate, keyUsages []x509.ExtKeyUsage) bool {
	usages := make([]x509.ExtKeyUsage, len(keyUsages))
	copy(usages, keyUsages)

	if len(chain) == 0 {
		return false
	}

	usagesRemaining := len(usages)

NextCert:
	for i := len(chain) - 1; i >= 0; i-- {
		cert := chain[i]
		if len(cert.ExtKeyUsage) == 0 && len(cert.UnknownExtKeyUsage) == 0 {
			continue
		}

		for _, usage := range cert.ExtKeyUsage {
			if usage == x509.ExtKeyUsageAny {
				continue NextCert
			}
		}

		const invalidUsage x509.ExtKeyUsage = -1

	NextRequestedUsage:
		for i, requestedUsage := range usages {
			if requestedUsage == invalidUsage {
				continue
			}

			for _, usage := range cert.ExtKeyUsage {
				if requestedUsage == usage {
					continue NextRequestedUsage
				} else if requestedUsage == x509.ExtKeyUsageServerAuth &&
					(usage == x509.ExtKeyUsageNetscapeServerGatedCrypto ||
						usage == x509.ExtKeyUsageMicrosoftServerGatedCrypto) {
					continue NextRequestedUsage
				}
			}

			usages[i] = invalidUsage
			usagesRemaining--
			if usagesRemaining == 0 {
				return false
			}
		}
	}

	return true
}

func (smc *SMCertificate) isValid(certType int, currentChain []*x509.Certificate, opts *SMVerifyOptions) error {
	c := (*x509.Certificate)(unsafe.Pointer(smc))
	now := opts.CurrentTime
	if now.IsZero() {
		now = time.Now()
	}
	if now.Before(c.NotBefore) || now.After(c.NotAfter) {
		return x509.CertificateInvalidError{Cert: c, Reason: x509.Expired}
	}

	if len(c.PermittedDNSDomains) > 0 {
		ok := false
		for _, domain := range c.PermittedDNSDomains {
			if opts.DNSName == domain ||
				(strings.HasSuffix(opts.DNSName, domain) &&
					len(opts.DNSName) >= 1+len(domain) &&
					opts.DNSName[len(opts.DNSName)-len(domain)-1] == '.') {
				ok = true
				break
			}
		}

		if !ok {
			return x509.CertificateInvalidError{Cert: c, Reason: x509.CANotAuthorizedForThisName}
		}
	}

	if certType == intermediateCertificate && (!c.BasicConstraintsValid || !c.IsCA) {
		return x509.CertificateInvalidError{Cert: c, Reason: x509.NotAuthorizedToSign}
	}

	if c.BasicConstraintsValid && c.MaxPathLen >= 0 {
		numIntermediates := len(currentChain) - 1
		if numIntermediates > c.MaxPathLen {
			return x509.CertificateInvalidError{Cert: c, Reason: x509.TooManyIntermediates}
		}
	}

	return nil
}

func (smc *SMCertificate) buildChains(cache map[int][][]*x509.Certificate, currentChain []*x509.Certificate, opts *SMVerifyOptions) (chains [][]*x509.Certificate, err error) {
	c := (*x509.Certificate)(unsafe.Pointer(smc))

	possibleRoots, failedRoot, rootErr := opts.Roots.findVerifiedParents(c)
	for _, rootNum := range possibleRoots {
		root := opts.Roots.certs[rootNum]
		err = root.isValid(rootCertificate, currentChain, opts)
		if err != nil {
			logger.Debugf("root valid fail. type: %v err: %v", rootCertificate, err)
			continue
		}
		chains = append(chains, appendToFreshChain(currentChain, root))
	}

	possibleIntermediates, failedIntermediate, intermediateErr := opts.Intermediates.findVerifiedParents(c)
nextIntermediate:
	for _, intermediateNum := range possibleIntermediates {
		intermediate := opts.Intermediates.certs[intermediateNum]
		for _, cert := range currentChain {
			if cert == (*x509.Certificate)(unsafe.Pointer(intermediate)) {
				continue nextIntermediate
			}
		}
		err = intermediate.isValid(intermediateCertificate, currentChain, opts)
		if err != nil {
			continue
		}
		var childChains [][]*x509.Certificate
		childChains, ok := cache[intermediateNum]
		if !ok {
			childChains, err = intermediate.buildChains(cache, appendToFreshChain(currentChain, intermediate), opts)
			cache[intermediateNum] = childChains
		}
		chains = append(chains, childChains...)
	}

	if len(chains) > 0 {
		err = nil
	}

	if len(chains) == 0 && err == nil {
		hintErr := rootErr
		hintCert := failedRoot
		if hintErr == nil {
			hintErr = intermediateErr
			hintCert = failedIntermediate
		}
		err = UnknownAuthorityError{c, hintErr, hintCert}
	}
	return
}

func (s *SMCertPool) findVerifiedParents(cert *x509.Certificate) (parents []int, errCert *x509.Certificate, err error) {
	smcert := (*SMCertificate)(unsafe.Pointer(cert))

	if s == nil {
		return
	}
	var candidates []int

	if len(cert.AuthorityKeyId) > 0 {
		candidates = s.bySubjectKeyId[string(cert.AuthorityKeyId)]
	}
	if len(candidates) == 0 {
		candidates = s.byName[string(cert.RawIssuer)]
	}

	for _, c := range candidates {
		if err = smcert.CheckSignatureFrom(s.certs[c]); err == nil {
			parents = append(parents, c)
		} else {
			errCert = (*x509.Certificate)(unsafe.Pointer(s.certs[c]))
		}
	}

	return
}

func appendToFreshChain(chain []*x509.Certificate, cert *SMCertificate) []*x509.Certificate {
	n := make([]*x509.Certificate, len(chain)+1)
	copy(n, chain)
	n[len(chain)] = (*x509.Certificate)(unsafe.Pointer(cert))
	return n
}

func (c *SMCertificate) CheckSignatureFrom(parent *SMCertificate) (err error) {
	if parent.Version == 3 && !parent.BasicConstraintsValid ||
		parent.BasicConstraintsValid && !parent.IsCA {
		return x509.ConstraintViolationError{}
	}

	if parent.KeyUsage != 0 && parent.KeyUsage&x509.KeyUsageCertSign == 0 {
		return x509.ConstraintViolationError{}
	}

	if parent.PublicKeyAlgorithm == x509.UnknownPublicKeyAlgorithm {
		return x509.ErrUnsupportedAlgorithm
	}

	return parent.CheckSignature(c.SignatureAlgorithm, c.RawTBSCertificate, c.Signature)
}

func (c *SMCertificate) CheckSignature(algo x509.SignatureAlgorithm, signed, signature []byte) (err error) {
	return checkSignature(algo, signed, signature, c.PublicKey)
}

func checkSignature(algo x509.SignatureAlgorithm, signed, signature []byte, publicKey crypto.PublicKey) (err error) {
	if algo == sx509.SM2WithSM3 {
		logger.Debug(">>>new version sm cert<<<")
		//h := SM.NewHash()
		//h.Write(signed)
		//digest := h.Sum(nil)
		digest := signed

		switch pub := publicKey.(type) {
		case *ecdsa.PublicKey:
			sm2verify_re, _ := hsm.VerifySM2(pub, signature, digest, nil)
			if !sm2verify_re {
				return errors.New("fail to verify signature")
			}

			return nil
		}
		return x509.ErrUnsupportedAlgorithm
	}
	return x509.ErrUnsupportedAlgorithm
}

func (s *SMCertPool) contains(cert *x509.Certificate) bool {
	if s == nil {
		return false
	}

	candidates := s.byName[string(cert.RawSubject)]
	for _, c := range candidates {
		if s.certs[c].Equal(cert) {
			return true
		}
	}

	return false
}

func (smc *SMCertificate) Equal(other *x509.Certificate) bool {
	return bytes.Equal(smc.Raw, other.Raw)
}
