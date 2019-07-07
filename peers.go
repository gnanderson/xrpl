package xrpl

/*
Copyright © 2019 Graham Anderson <graham@grahamanderson.scot>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
)

const (
	unstable = "unknown"
	insane   = "insane"
)

var currentVer = semver.Must(semver.NewVersion("1.2.4"))

// DefaultStabilityChecker is the packages own opintionated check function for use
// with a peers StableWith method.
var DefaultStabilityChecker = &PeerList{}

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
// is unrecognised e.g. not 'rippled-x.x.x' it returns a nil version and error
func (p Peer) SemVer() (*semver.Version, error) {
	v := strings.TrimLeft(p.Version, "rippled-")

	return semver.NewVersion(v)
}

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
		Peers  []*Peer `json:"peers"`
		Status string  `json:"status"`
	} `json:"result"`
}

// Peers returns the list of conntected peers in the peer list
func (pl *PeerList) Peers() []*Peer {
	return pl.Result.Peers
}

// Stable returns a list of peers that are not reporting unknown or insane
// values. If you need to examine the state further you can use custom stability
// rules using the Peer.StableWith() method
func (pl *PeerList) Stable() []*Peer {
	braw := make([]*Peer, 0)
	for _, peer := range pl.Peers() {
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
func (pl *PeerList) Unstable() []*Peer {
	b0rked := make([]*Peer, 0)
	for _, peer := range pl.Peers() {
		if peer.Sanity == unstable || peer.Sanity == insane {
			b0rked = append(b0rked, peer)
		}
	}

	return b0rked
}

// Check is an opinionated view of the peers stability based on the reported
// sanity field, the version and the connected uptime. This checker returns
// true only if the peer is sane, is a recent version of rippled, and has
// been connected long enough to decide on it's sanity.
func (pl *PeerList) Check(p Peer) bool {
	peerVer, err := p.SemVer()
	if err != nil {
		log.Println("peer version:", err)
		return false
	}

	minCheckAge := 30 * time.Minute
	checkPeer := p.Uptime > int(minCheckAge.Seconds())
	sane := p.Sanity != unstable && p.Sanity != insane

	if checkPeer {
		return sane && !versionTooOld(currentVer, peerVer)
	}

	return true
}

// Anonymise (randomise) the IP's of a peerlist - for testing / CI purposes...
// public keys are obviously in the public domain but lets be polite about the IP's
func (pl *PeerList) Anonymise() {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, peer := range pl.Peers() {
		blocks := []string{}
		for i := 0; i < 4; i++ {
			number := rand.Intn(255)
			blocks = append(blocks, strconv.Itoa(number))
		}

		addr := fmt.Sprintf("%s:%d", strings.Join(blocks, "."), rand.Intn(60000))
		peer.Address = addr
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	if err := enc.Encode(pl); err != nil {
		fmt.Println(err)
	}
}

// UnmarshalPeers is a convenience function to marshal a JSON peer list into
// a *PeerList
func UnmarshalPeers(peerList string) (*PeerList, error) {
	list := &PeerList{}
	err := json.Unmarshal([]byte(peerList), list)

	return list, err
}
