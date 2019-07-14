package xrpl

import (
	"io"

	"github.com/BurntSushi/toml"
)

// Toml is a representation of /.well-known/srp-ledger.toml
type Toml struct {
	METADATA   *TomlMetadata
	VALIDATORS []*TomlValidator
	PRINCIPLES []*TomlPrinciple
	SERVERS    []*TomlServer
}

// Encode writes the toml to the io.Writer provided
func (t *Toml) Encode(w io.Writer) error {
	enc := toml.NewEncoder(w)

	return enc.Encode(t)
}

// TomlMetadata represents the [[METADATA]] entry
type TomlMetadata struct {
	Modified string
}

// TomlValidator represents an entry in the [[VALIDATORS]] toml list
type TomlValidator struct {
	PubKey        string
	Network       string
	OwnerCountry  string
	ServerCountry string
	UNL           string
}

// TomlPrinciple represents an dentry in the [[PRINCIPLES]] toml list
type TomlPrinciple struct {
	Name  string
	email string
}

// TomlServer represents an dentry in the [[SERVERS]] toml list
type TomlServer struct {
	Peer    string
	Network string
	Port    int
}
