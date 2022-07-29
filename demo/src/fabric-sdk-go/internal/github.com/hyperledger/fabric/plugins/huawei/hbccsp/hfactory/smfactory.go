package hfactory

import (
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/plugins/huawei/hbccsp/internal/hsm"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
)

const (
	// SoftwareBasedFactoryName is the name of the factory of the software-based BCCSP implementation
	SMSoftwareBasedFactoryName = "SM"
)

// SMFactory is the factory of the software-based BCCSP.
type SMFactory struct{}

// Name returns the name of this factory
func (f *SMFactory) Name() string {
	return SMSoftwareBasedFactoryName
}

// Get returns an instance of BCCSP using Opts.
func (f *SMFactory) Get(smOpts *SmOpts) (bccsp.BCCSP, error) {
	// Validate arguments
	if smOpts == nil {
		return nil, errors.New("Invalid config. It must not be nil.")
	}

	var ks bccsp.KeyStore
	if smOpts.Ephemeral == true {
		ks = hsm.NewDummyKeyStore()
	} else if smOpts.FileKeystore != nil {
		fks, err := hsm.NewFileBasedKeyStore(nil, smOpts.FileKeystore.KeyStorePath, false)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to initialize software key store")
		}
		ks = fks
	} else {
		// Default to DummyKeystore
		ks = hsm.NewDummyKeyStore()
	}

	return hsm.New(smOpts.SecLevel, smOpts.HashFamily, ks)
}

// SMOpts contains options for the SMFactory
type SmOpts struct {
	// Default algorithms when not specified (Deprecated?)
	SecLevel   int    `mapstructure:"security" json:"security" yaml:"Security"`
	HashFamily string `mapstructure:"hash" json:"hash" yaml:"Hash"`

	// Keystore Options
	Ephemeral     bool                 `mapstructure:"tempkeys,omitempty" json:"tempkeys,omitempty"`
	FileKeystore  *SMFileKeystoreOpts  `mapstructure:"filekeystore,omitempty" json:"filekeystore,omitempty" yaml:"FileKeyStore"`
	DummyKeystore *SMDummyKeystoreOpts `mapstructure:"dummykeystore,omitempty" json:"dummykeystore,omitempty"`
}

// Pluggable Keystores, could add JKS, P12, etc..
type SMFileKeystoreOpts struct {
	KeyStorePath string `mapstructure:"keystore" yaml:"KeyStore"`
}

type SMDummyKeystoreOpts struct{}
