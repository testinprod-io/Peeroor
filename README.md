# Peeroor

This project is p2p network manager for Ethereum execution client nodes. It connects nodes via their RPC endpoints, forms a full-mesh network, and maintains reliable peer connections by periodically checking and re-adding missing connections.

It uses `admin_addPeer` to add peers to the nodes, and requires `admin` rpc namespace to be enabled. 

## Configuration

Create a `config.yaml` file in the project root. Example:

```yaml
rpcs:
  rpc1: "http://localhost:8545"
  rpc2: "http://localhost:8546"
  rpc3: "http://localhost:8547"
  rpc4: "http://localhost:8548"

networks:
  network1:
    - rpc1
    - rpc2
  network2:
    - rpc3
    - rpc4
```