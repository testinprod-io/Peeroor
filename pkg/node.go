package pkg

import (
	"fmt"
	"github.com/ethereum/go-ethereum/rpc"
	"sync"
	"time"
)

// Node represents an Ethereum execution client node.
type Node struct {
	Name     string // The key name from the config (e.g., "rpc1")
	Endpoint string

	Enode          string
	LastPeerUpdate time.Time // When the ConnectedPeers map was last updated
	mu             sync.RWMutex
	Peers          map[string]bool // peers (enode string -> connected setate)
}

// GetEnode retrieves the enode string from the node using admin_nodeInfo.
func (n *Node) GetEnode() (string, error) {
	var result struct {
		Enode string `json:"enode"`
	}

	client, err := rpc.Dial(n.Endpoint)
	if err != nil {
		return "", fmt.Errorf("Failed to connect to %s (%s): %v", n.Name, n.Endpoint, err)
	}
	if err := client.Call(&result, "admin_nodeInfo"); err != nil {
		return "", err
	}
	return result.Enode, nil
}

// GetConnectedPeersMap fetches the current list of connected peers using admin_peers.
func (n *Node) GetConnectedPeersMap() (map[string]bool, error) {
	var peers []struct {
		Enode string `json:"enode"`
	}
	client, err := rpc.Dial(n.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to %s (%s): %v", n.Name, n.Endpoint, err)
	}
	if err := client.Call(&peers, "admin_peers"); err != nil {
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
	var result interface{}

	client, err := rpc.Dial(n.Endpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to %s (%s): %v", n.Name, n.Endpoint, err)
	}
	err = client.Call(&result, "admin_addPeer", peerEnode)

	if err != nil {
		return err
	}

	n.mu.Lock()
	n.Peers[peerEnode] = true
	n.mu.Unlock()
	return nil
}

// RefreshPeers updates the cached Peers map with actual connection status.
// It only refreshes if the cache is expired.
func (n *Node) RefreshPeers() error {
	n.mu.RLock()
	if n.Peers != nil && time.Since(n.LastPeerUpdate) < connectedPeersExpiration {
		n.mu.RUnlock()
		return nil // cache is still valid
	}
	n.mu.RUnlock()

	actual, err := n.GetConnectedPeersMap()
	if err != nil {
		return err
	}

	n.mu.Lock()
	defer n.mu.Unlock()
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
