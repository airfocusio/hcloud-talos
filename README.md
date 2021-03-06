# hcloud-talos

## Commands

This CLI tool provides an easy way to manage [Talos](https://talos.dev/) powered [Kubernetes](https://kubernetes.io/) clusters on the [Hetzner Cloud](https://www.hetzner.com/cloud). Bootstrapping a new cluster performs the following steps:

* Create private network `10.0.0.0/16` for inter-node communication
* Create placement group to ensure nodes to not run on the same physical machine
* Create load balancer to access the controlplane nodes Kubernetes API server (port `6443`) or Talos API server (port `50000`)
* Create firewall rules to block access to nodes from outside of the private network
* Create first controlplane node
* Install [Hetzner Cloud Controller Manger](https://github.com/hetznercloud/hcloud-cloud-controller-manager)
* Install [Hetzner CSI Driver](https://github.com/hetznercloud/csi-driver)

## Usage

```bash
# ATTENTION: this folder will contain all crucial files and they must be stored somewhere secure!
mkdir my-cluster
cd my-cluster

export HCLOUD_TOKEN=...
# bootstrap cluster
hcloud-talos bootstrap-cluster my-cluster controlplane-1

# add more nodes
hcloud-talos add-node controlplane-2 --controlplane
hcloud-talos add-node worker-1
```
