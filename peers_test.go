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
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

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

var verTests = []struct {
	min      string
	compare  string
	expected bool
}{
	{"1.2.4", "1.2.3", true},
	{"1.2.4", "1.3.0", false},
	{"1.2.4", "1.2.4-beta1", true},
	{"1.2.4", "1.3.0-rc1", false},
}

func TestVersions(t *testing.T) {
	tf := func(min, compare string, expected bool) {
		MinVersion = semver.Must(semver.NewVersion(min))

		compareP := &Peer{Version: compare}
		lessThan := compareP.TooOld()

		if lessThan != expected {
			t.Errorf("Version problem - min: '%s' compare:  %s lessthan: %t", min, compare, lessThan)
		}
	}

	for _, tt := range verTests {
		t.Run(tt.min, func(t *testing.T) {
			tf(tt.min, tt.compare, tt.expected)
		})
	}
}
