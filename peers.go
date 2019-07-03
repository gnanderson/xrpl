package xrpl

import (
	"encoding/json"
	"log"
	"net"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
)

const (
	unstable = "unknown"
	insane   = "insane"
)

var currentVer = semver.Must(semver.NewVersion("1.2.4"))
var prevMaxVer = semver.Must(semver.NewVersion("1.1.3"))

// Peer defines a validator/stock peer node
type Peer struct {
	Address         string `json:"address"`
	CompleteLedgers string `json:"complete_ledgers,omitempty"`
	Inbound         bool   `json:"inbound,omitempty"`
	Latency         int    `json:"latency"`
	Ledger          string `json:"ledger,omitempty"`
	Load            int    `json:"load"`
	PublicKey       string `json:"public_key"`
	Uptime          int    `json:"uptime"`
	Version         string `json:"version"`
	Sanity          string `json:"sanity,omitempty"`
	Cluster         bool   `json:"cluster,omitempty"`
}

// IP returns the network IP address of the peer
func (p Peer) IP() net.IP {
	host, _, err := net.SplitHostPort(p.Address)
	if err != nil {
		log.Fatal(err)

	}

	return net.ParseIP(host)
}

// StableWith checks the stability of a node with a StabilityChecker
func (p Peer) StableWith(checker StabilityChecker) bool {
	return checker.Check(p)
}

// SemVer returns the semantic version of the node software, if the version string
// is unrecognised e.g. not 'rippled-X' it returns a nil version and error
func (p Peer) SemVer() (*semver.Version, error) {
	v := strings.TrimLeft(p.Version, "rippled-")

	return semver.NewVersion(v)
}

// only accept versions that are one patch behind
func versionTooOld(current, compare *semver.Version) bool {
	if compare.LessThan(*current) {
		compare.BumpPatch()
		return compare.LessThan(*current)
	}

	return false
}

// StabilityChecker is the interface that a peer can consumer to check the
// stability of a peer against some custom rules.
type StabilityChecker interface {
	Check(peer Peer) bool
}

// PeerList represents the output from the 'peers' admin commnad
type PeerList struct {
	Result struct {
		Peers  []Peer `json:"peers"`
		Status string `json:"status"`
	} `json:"result"`
}

// Peers returns the list of conntected peers in the peer list
func (pl *PeerList) Peers() []Peer {
	return pl.Result.Peers
}

// Stable returns a list of peers that are not reporting unknown or insane
// values. If you need to examine the state further you can use custom stability
// rules using the Peer.StableWith() method
func (pl *PeerList) Stable() []Peer {
	braw := make([]Peer, 0)
	for _, peer := range pl.Peers() {
		// if peer sanity is the zero value
		if peer.Sanity == "" {
			braw = append(braw, peer)
		}
	}
	return braw
}

// Unstable returns unkown or insane peers, typically you would want to
// further check individual peers for uptime and/or old rippled versions
// before you take action. You can use the Peer.StableWith() method to test the
// peer further.
func (pl *PeerList) Unstable() []Peer {
	b0rked := make([]Peer, 0)
	for _, peer := range pl.Peers() {
		if peer.Sanity == unstable || peer.Sanity == insane {
			b0rked = append(b0rked, peer)
		}
	}

	return b0rked
}

// Check is an opinionated view of the peers stability based on the reported
// sanity field, the version and the connected uptime.
func (pl *PeerList) Check(p Peer) bool {
	sane := p.Sanity != unstable && p.Sanity != insane
	peerVer, err := p.SemVer()
	maxAge := 30 * time.Minute
	isPastMaxAge := p.Uptime > int(maxAge.Seconds())

	if err != nil {
		log.Fatal(err)
	}

	switch {
	case !sane && isPastMaxAge:
		return false
	case sane && versionTooOld(currentVer, peerVer):
		return false
	}

	return true
}

func unmarshalPeers(peerList string) (*PeerList, error) {
	list := &PeerList{}
	err := json.Unmarshal([]byte(peerList), list)

	return list, err
}
