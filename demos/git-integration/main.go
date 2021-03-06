package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/brynbellomy/klog"

	rw "redwood.dev"
	"redwood.dev/ctx"
	"redwood.dev/types"
	"redwood.dev/utils"
)

type app struct {
	ctx.Context
}

func main() {
	flagset := flag.NewFlagSet("", flag.ContinueOnError)
	klog.InitFlags(flagset)
	flagset.Set("logtostderr", "true")
	flagset.Set("v", "2")
	klog.SetFormatter(&klog.FmtConstWidth{
		FileNameCharWidth: 24,
		UseColor:          true,
	})

	// Make two Go hosts that will communicate with one another over libp2p
	var host1 rw.Host
	var host2 rw.Host

	err := host1.Start()
	if err != nil {
		panic(err)
	}
	defer host1.Close()

	err = host2.Start()
	if err != nil {
		panic(err)
	}
	defer host2.Close()

	// Connect the two peers using libp2p
	type libp2pPeerIDer interface {
		Libp2pPeerID() string
	}
	libp2pTransport := host1.Transport("libp2p").(libp2pPeerIDer)
	host2.AddPeer(rw.PeerDialInfo{"libp2p", "/ip4/0.0.0.0/tcp/21231/p2p/" + libp2pTransport.Libp2pPeerID()})

	time.Sleep(2 * time.Second)

	// Both consumers subscribe to the StateURI
	ctx, _ := context.WithTimeout(context.Background(), 120*time.Second)
	go func() {
		_, err := host2.Subscribe(ctx, "somegitprovider.org/gitdemo", rw.SubscriptionType_Txs, nil, nil)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		_, err := host1.Subscribe(ctx, "somegitprovider.org/gitdemo", rw.SubscriptionType_Txs, nil, nil)
		if err != nil {
			panic(err)
		}
	}()

	sendTxs(host1, host2)

	<-utils.AwaitInterrupt()
}

func sendTxs(host1, host2 rw.Host) {
	// Before sending any transactions, we upload some resources we're going to need
	// into the RefStore of the node.  These resources can be referred to in the state
	// tree by their hash.
	indexHTML, err := os.Open("./repo/index.html")
	if err != nil {
		panic(err)
	}
	indexHTMLHash, _, err := host1.AddRef(indexHTML)
	if err != nil {
		panic(err)
	}
	scriptJS, err := os.Open("./repo/script.js")
	if err != nil {
		panic(err)
	}
	scriptJSHash, _, err := host1.AddRef(scriptJS)
	if err != nil {
		panic(err)
	}
	readme, err := os.Open("./repo/README.md")
	if err != nil {
		panic(err)
	}
	readmeHash, _, err := host1.AddRef(readme)
	if err != nil {
		panic(err)
	}
	redwoodJpg, err := os.Open("./repo/redwood.jpg")
	if err != nil {
		panic(err)
	}
	redwoodJpgHash, _, err := host1.AddRef(redwoodJpg)
	if err != nil {
		panic(err)
	}

	identities1, err := host1.Identities()
	if err != nil {
		panic(err)
	}
	identities2, err := host2.Identities()
	if err != nil {
		panic(err)
	}
	host1Addr := identities1[0].Address()
	host2Addr := identities2[0].Address()

	// These are just convenience utils
	hostsByAddress := map[types.Address]rw.Host{
		host1Addr: host1,
		host2Addr: host2,
	}

	sendTx := func(tx rw.Tx) {
		host := hostsByAddress[tx.From]
		err := host.SendTx(context.Background(), tx)
		if err != nil {
			host.Errorf("%+v", err)
		}
	}

	// If you alter the contents of the ./repo subdirectory, you'll need to determine the
	// git commit hash of the first commit again, and then tweak these variables.  Otherwise,
	// you'll get a "bad object" error from git.
	commit1Hash := "2d4518de34a9583d61b32c9bf3b4cf0bdc1c8734"
	commit1Timestamp := "2020-05-26T16:42:24-05:00"

	commit1RepoTxID, err := types.IDFromHex(commit1Hash)
	if err != nil {
		panic(err)
	}

	//
	// Setup our git repo's state tree.
	//
	var (
		// The "gitdemo" channel contains:
		//   - A link to the current worktree so that we can browse it like a regular website.
		//   - All of the commit data that Git expects.  The files are stored under a "files" key,
		//         and other important metadata are stored under the other keys.
		//   - A mapping of refs (usually, branches) to commit hashes.
		//   - A permissions validator that allows anyone to write to the repo but tries to keep
		//         people from writing to the wrong keys.
		genesisDemo = rw.Tx{
			ID:       rw.GenesisTxID,
			Parents:  []types.ID{},
			From:     host1Addr,
			StateURI: "somegitprovider.org/gitdemo",
			Patches: []rw.Patch{
				mustParsePatch(` = {
                    "demo": {
                        "Content-Type": "link",
                        "value": "state:somegitprovider.org/gitdemo/refs/heads/master/worktree"
                    },
					"Merge-Type": {
						"Content-Type": "resolver/dumb",
						"value": {}
					},
					"Validator": {
						"Content-Type": "validator/permissions",
						"value": {
							"96216849c49358b10257cb55b28ea603c874b05e": {
								"^.*$": {
									"write": true
								}
							},
                            "*": {
                                "^\\.refs\\..*": {
                                    "write": true
                                },
                                "^\\.commits\\.[a-f0-9]+\\.parents": {
                                    "write": true
                                },
                                "^\\.commits\\.[a-f0-9]+\\.message": {
                                    "write": true
                                },
                                "^\\.commits\\.[a-f0-9]+\\.timestamp": {
                                    "write": true
                                },
                                "^\\.commits\\.[a-f0-9]+\\.author": {
                                    "write": true
                                },
                                "^\\.commits\\.[a-f0-9]+\\.committer": {
                                    "write": true
                                },
                                "^\\.commits\\.[a-f0-9]+\\.files": {
                                    "write": true
                                }
                            }
						}
					},
                    "refs": {
                        "heads": {}
                    },
                    "commits": {}
				}`),
			},
		}

		// Finally, we submit two transactions (one to "git" and one to "git-reflog") that simulate what
		// would happen if a user were to push their first commit to the repo.  The repo is now able to be
		// cloned using the command "git clone redwood://localhost:21232@somegitprovider.org/gitdemo"
		commit1Repo = rw.Tx{
			ID:         commit1RepoTxID,
			Parents:    []types.ID{genesisDemo.ID},
			From:       host1Addr,
			StateURI:   "somegitprovider.org/gitdemo",
			Checkpoint: true,
			Patches: []rw.Patch{
				mustParsePatch(`.commits.` + commit1Hash + ` = {
                    "message": "First commit\n",
                    "timestamp": "` + commit1Timestamp + `",
                    "author": {
                        "name": "Bryn Bellomy",
                        "email": "bryn.bellomy@gmail.com",
                        "timestamp": "` + commit1Timestamp + `"
                    },
                    "committer": {
                        "name": "Bryn Bellomy",
                        "email": "bryn.bellomy@gmail.com",
                        "timestamp": "` + commit1Timestamp + `"
                    },
                    "files": {
                        "README.md": {
                            "Content-Type": "link",
                            "mode": 33188,
                            "value": "ref:sha1:` + readmeHash.Hex() + `"
                        },
                        "redwood.jpg": {
                            "Content-Type": "link",
                            "mode": 33188,
                            "value": "ref:sha1:` + redwoodJpgHash.Hex() + `"
                        },
                        "index.html": {
                            "Content-Type": "link",
                            "mode": 33188,
                            "value": "ref:sha1:` + indexHTMLHash.Hex() + `"
                        },
                        "script.js": {
                            "Content-Type": "link",
                            "mode": 33188,
                            "value": "ref:sha1:` + scriptJSHash.Hex() + `"
                        }
                    }
                }`),
				mustParsePatch(`.refs.heads.master = {
                    "HEAD": "` + commit1Hash + `",
                    "worktree": {
                        "Content-Type": "link",
                        "value": "state:somegitprovider.org/gitdemo/commits/` + commit1Hash + `/files"
                    }
                }`),
			},
		}
	)

	sendTx(genesisDemo)
	sendTx(commit1Repo)
}

func mustParsePatch(s string) rw.Patch {
	p, err := rw.ParsePatch([]byte(s))
	if err != nil {
		panic(err.Error() + ": " + s)
	}
	return p
}
