# hcloud-talos

## Commands

This CLI tool provides an easy way to manage [Talos](https://talos.dev/) powered [Kubernetes](https://kubernetes.io/) clustes on the [Hetzner Cloud](https://www.hetzner.com/cloud).

* `bootstrap-cluster`
    * Create a private network `10.0.0.0/8` for inter-node communication
    * Create a placement group to ensure nodes to not run on the same physical machine
    * Create a firewall that blocks all incoming traffic (except ICMP)
    * Create a load balancer to access the controlplane nodes via Kubernetes API server (port `6443`) or Talos API server (port `50000`)
    * Create a load balancer to access the nodes ingress (port `80` -> `30080`, `443` -> `30433`) with proxy protocol
    * Create a first controlplane node running Talos
    * Install [Hetzner Cloud Controller Manger](https://github.com/hetznercloud/hcloud-cloud-controller-manager)
    * Install [Hetzner CSI Driver](https://github.com/hetznercloud/csi-driver)
* `add-node`
    * Create an additional node running Talos, either controlplane or worker
* `destroy-cluster`
    * Remove all Hetzner resources (by label)

## Usage

```bash
mkdir my-cluster
cd my-cluster
export HCLOUD_TOKEN=...
# bootstrap cluster
hcloud-talos bootstrap-cluster --cluster-name=my-cluster --node-name=controlplane-01 --force

# add more nodes
hcloud-talos add-node --cluster-name=cluster --node-name=controlplane-02 --controlplane
hcloud-talos add-node --node-name=worker-01
```

## Development

```bash
export HCLOUD_TOKEN=...
go run . --dir=test --verbose command [args]
```
