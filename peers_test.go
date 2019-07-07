package xrpl

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/coreos/go-semver/semver"
)

func peerList() *PeerList {
	f, err := os.Open("peers_anon.json")
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	peerList, err := UnmarshalPeers(string(bytes))
	if err != nil {
		log.Fatal(err)
	}

	return peerList
}

func TestPeersUnmarshal(t *testing.T) {
	for _, peer := range peerList().Peers() {
		if peer.Address == "" {
			t.Fatal(peer)
		}

		if peer.IP() == nil {
			t.Fatalf("expected IP address parsed incorrectlly: '%s'", peer.Address)
		}
	}
}

func TestPeersGetUnstable(t *testing.T) {
	pl := peerList()
	for _, peer := range pl.Unstable() {
		if peer.StableWith(pl) {
			t.Fatal(peer)
		}
	}

	if len(pl.Unstable()) < 3 {
		t.Fatalf("expected %d unstable peers, got %d", 3, len(pl.Unstable()))
	}
}

func TestVersionTooOld(t *testing.T) {
	compare := semver.New("1.1.3")
	current := semver.New("1.2.4")

	if !versionTooOld(current, compare) {
		t.Fatalf("A (%s) is not less than B (%s)", compare, current)
	}
}
