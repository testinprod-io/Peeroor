package Peeroor

import (
	"github.com/ethereum/go-ethereum/rpc"
	"log"
	"time"
)

// Node represents an Ethereum execution client node.
type Node struct {
	Name           string // The key name from the config (e.g., "rpc1")
	Endpoint       string
	Client         *rpc.Client
	Enode          string
	LastPeerUpdate time.Time // When the ConnectedPeers map was last updated

	Peers map[string]bool // peers (enode string -> connected setate)
}

// GetEnode retrieves the enode string from the node using admin_nodeInfo.
func (n *Node) GetEnode() (string, error) {
	var result struct {
		Enode string `json:"enode"`
	}
	if err := n.Client.Call(&result, "admin_nodeInfo"); err != nil {
		return "", err
	}
	return result.Enode, nil
}

// GetConnectedPeersMap fetches the current list of connected peers using admin_peers.
func (n *Node) GetConnectedPeersMap() (map[string]bool, error) {
	var peers []struct {
		Enode string `json:"enode"`
	}
	if err := n.Client.Call(&peers, "admin_peers"); err != nil {
		return nil, err
	}
	peerMap := make(map[string]bool, len(peers))
	for _, peer := range peers {
		peerMap[peer.Enode] = true
	}
	return peerMap, nil
}

// AddPeer calls admin_addPeer to add the given peer.

func (n *Node) AddPeer(peerEnode string) error {
	var result bool

	// First, try using admin_addTrustedPeer.
	err := n.Client.Call(&result, "admin_addTrustedPeer", peerEnode)
	if err != nil {
		// Log the error and fall back to admin_addPeer.
		log.Printf("Node %s: admin_addTrustedPeer failed: %v, falling back to admin_addPeer", n.Name, err)
		err = n.Client.Call(&result, "admin_addPeer", peerEnode)
		if err != nil {
			return err
		}
	}
	n.Peers[peerEnode] = true
	return nil
}

// Reconnect attempts to re-establish the rpc.Client connection.
func (n *Node) Reconnect() error {
	client, err := rpc.Dial(n.Endpoint)
	if err != nil {
		return err
	}
	n.Client = client
	return nil
}

// RefreshPeers updates the cached Peers map with actual connection status.
// It only refreshes if the cache is expired.
func (n *Node) RefreshPeers() error {
	if n.Peers != nil && time.Since(n.LastPeerUpdate) < connectedPeersExpiration {
		return nil // cache is still valid
	}

	actual, err := n.GetConnectedPeersMap()
	if err != nil {
		return err
	}

	// For each desired peer in n.Peers, update its status based on actual connections.
	for peer := range n.Peers {
		if actual[peer] {
			n.Peers[peer] = true
		} else {
			n.Peers[peer] = false
		}
	}

	n.LastPeerUpdate = time.Now()
	return nil
}
