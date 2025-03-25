package Peeroor

import (
	"github.com/ethereum/go-ethereum/rpc"
	"log"
	"sync"
	"time"
)

// Network represents a network with a collection of nodes.
type Network struct {
	Name   string
	Nodes  []*Node
	Ticker *time.Ticker
	Stop   chan bool
}

// NewNetwork creates and initializes a Network from the given rpcKeys and config map.
func NewNetwork(name string, rpcKeys []string, rpcMap map[string]string) *Network {
	network := &Network{
		Name:   name,
		Nodes:  make([]*Node, 0, len(rpcKeys)),
		Ticker: time.NewTicker(30 * time.Second),
		Stop:   make(chan bool),
	}

	// Connect to each node defined in this network.
	for _, key := range rpcKeys {
		url, ok := rpcMap[key]
		if !ok {
			log.Printf("Network %s: RPC key %s not found in config, skipping", name, key)
			continue
		}
		client, err := rpc.Dial(url)
		if err != nil {
			log.Printf("Network %s: Failed to connect to %s (%s): %v", name, key, url, err)
			continue
		}
		node := &Node{
			Name:     key,
			Endpoint: url,
			Client:   client,
			Peers:    make(map[string]bool),
		}
		enode, err := node.GetEnode()
		if err != nil {
			log.Printf("Network %s: Failed to get enode for %s (%s): %v", name, key, url, err)
			continue
		}
		node.Enode = enode
		network.Nodes = append(network.Nodes, node)
		log.Printf("Network %s: Connected node %s with enode %s", name, key, enode)
	}

	// Set each node's desired peers (all other nodes in this network).
	for i, node := range network.Nodes {
		for j, other := range network.Nodes {
			if i == j {
				continue
			}
			node.Peers[other.Enode] = false
		}
	}

	// Build the initial full-mesh network.
	network.UpdatePeers()

	return network
}

// Maintain runs the ticker for the network to periodically recheck and re-establish missing connections.
func (network *Network) Maintain() {
	log.Printf("Network %s: Starting maintenance ticker", network.Name)
	for {
		select {
		case <-network.Ticker.C:
			network.UpdatePeers()
		case <-network.Stop:
			log.Printf("Network %s: Stopping maintenance ticker", network.Name)
			return
		}
	}
}

func (network *Network) UpdatePeers() {
	var wg sync.WaitGroup

	n := len(network.Nodes)

	log.Printf("Network %s: Rechecking peer connections", network.Name)

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			// Check node[i] -> node[j]
			wg.Add(1)
			go func(node, peerNode *Node) {
				defer wg.Done()
				if err := node.RefreshPeers(); err != nil {
					log.Printf("Network %s: Error refreshing peers for node %s: %v", network.Name, node.Name, err)
					if err := node.Reconnect(); err != nil {
						log.Printf("Network %s: Failed to reconnect node %s: %v", network.Name, node.Name, err)
					}
					return
				}
				if !node.Peers[peerNode.Enode] {
					if err := node.AddPeer(peerNode.Enode); err != nil {
						log.Printf("Network %s: Error re-adding peer %s to node %s: %v", network.Name, peerNode.Enode, node.Name, err)
					} else {
						log.Printf("Network %s: Node %s re-added peer %s", network.Name, node.Name, peerNode.Enode)
					}
				}
			}(network.Nodes[i], network.Nodes[j])

			// Check node[j] -> node[i]
			wg.Add(1)
			go func(node, peerNode *Node) {
				defer wg.Done()
				if err := node.RefreshPeers(); err != nil {
					log.Printf("Network %s: Error refreshing peers for node %s: %v", network.Name, node.Name, err)
					if err := node.Reconnect(); err != nil {
						log.Printf("Network %s: Failed to reconnect node %s: %v", network.Name, node.Name, err)
					}
					return
				}
				if !node.Peers[peerNode.Enode] {
					if err := node.AddPeer(peerNode.Enode); err != nil {
						log.Printf("Network %s: Error re-adding peer %s to node %s: %v", network.Name, peerNode.Enode, node.Name, err)
					} else {
						log.Printf("Network %s: Node %s re-added peer %s", network.Name, node.Name, peerNode.Enode)
					}
				}
			}(network.Nodes[j], network.Nodes[i])
		}
	}
	wg.Wait()
}
