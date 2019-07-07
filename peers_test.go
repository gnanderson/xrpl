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
