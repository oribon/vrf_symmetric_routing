
Start with a cluster that has at least 2 workers, on each of them there's an interface (ens8) belonging to a network isolated from the primary one (br-ex), and another machine (vm) connected to the same network as the ens8 interface. Here we assume it looks like this:
```
       ┌───────────────────────────────────────────────────────────────────────────────┐           
       │  OpenShift Cluster                                                            │           
       │                                                                               │           
       │                                                                               │           
       │                                                                               │           
       │                                                                               │           
       │                                                                               │           
       │         ┌──────────────────────┐            ┌──────────────────────┐          │           
       │         │ worker0.lab.com      │            │ worker1.lab.com      │          │           
       │         │                      │            │                      │          │           
       │         │                      │            │                      │          │           
       │         │                      │            │                      │          │           
       │         │                      │            │                      │          │           
       │         │                      │            │                      │          │           
       │         │  br-ex        ens8   │            │  br-ex        ens8   │          │           
       │         │  ┌───┐        ┌───┐  │            │  ┌───┐        ┌───┐  │          │           
       │         └──┼───┼────────┼───┼──┘            └──┼───┼────────┼───┼──┘          │           
       │            │   │        │   │                  │   │        │   │             │           
       │            │   │        │   │                  │   │        │   │             │           
       │            └───┘        └───┘                  └───┘        └───┘             │           
       │                     192.168.100.246                     192.168.100.220       │           
       │                                                                               │           
       │                                                                               │           
       │                                                                               │           
       └───────────────────────────────────────────────────────────────────────────────┘           
                                                                                                   
                                                                                                   
                                                                                                   
                                                                                                   
                                       192.168.100.100                                             
                                            ┌───┐                                                  
                                            │   │                                                  
                                   ┌────────│───│────────┐                                         
                                   │        │   │        │                                         
                                   │        └───┘        │                                         
                                   │                     │                                         
                                   │                     │                                         
                                   │                     │                                         
                                   │                     │                                         
                                   │                     │                                         
                                   │External Host        │                                         
                                   └─────────────────────┘                                         
```
**From now on we assume that all of the commands are ran on the external host.**
We also assume the external host has podman/docker installed.

* Label the two workers with `vrf: true`:
```
kubectl label node worker0.lab.com vrf=true
kubectl label node worker1.lab.com vrf=true
```

* Ensure cluster uses LGW + global ipForwarding:
```
k edit network.operator.openshift.io cluster
```

```yaml
  defaultNetwork:
    ovnKubernetesConfig:
      gatewayConfig:
        ipForwarding: Global
        routingViaHost: true
```
If it did not, wait for the pods in the `openshift-ovn-kubernetes` namespace to rollout.

* Install the MetalLB and NMstate operators:
```
kubectl apply -f install_operators.yaml
```
* Wait for the CSVs in the `openshift-metallb-system`,`openshift-nmstate` namespaces to be in "Succeeded" phase.
* Install MetalLB and NMState:
```yaml
apiVersion: metallb.io/v1beta1
kind: MetalLB
metadata:
  name: metallb
  namespace: openshift-metallb-system
---
apiVersion: nmstate.io/v1
kind: NMState
metadata:
  name: nmstate
```
* Wait for the pods in the `openshift-metallb-system`,`openshift-nmstate` namespaces to reach a running state.

* Fill `setup/conf.yaml` according to your environment, example:
```yaml
intf: "ens8"
worker0:
  nodeName: "worker0.lab.com"
  ip: "192.168.100.246"
worker1:
  nodeName: "worker1.lab.com"
  ip: "192.168.100.220"
externalHostIP: "192.168.100.100"
secondaryNetGW: "192.168.100.1"
```

* Generate the relevant configuration resources by running:
```
cd setup
go run main.go
```
* Apply the configurations that were generated:
```
kubectl apply -f vrf-nncps.yaml
kubectl apply -f metallb.yaml
```
* Run FRR on the external host:
```
podman run -v ./frr:/etc/frr --privileged --net=host -d --name=frr quay.io/frrouting/frr:9.1.0
```
After about 2 minutes the external host should be peered with the workers:
```
podman exec frr vtysh -c 'show bgp summary established'

IPv4 Unicast Summary (VRF default):
BGP router identifier 192.168.100.100, local AS number 64200 vrf-id 0
BGP table version 59
RIB entries 1, using 96 bytes of memory
Peers 2, using 26 KiB of memory

Neighbor        V         AS   MsgRcvd   MsgSent   TblVer  InQ OutQ  Up/Down  State/PfxRcd   PfxSnt  Desc
192.168.100.220 4      64100         1         1        0    0    0       3m             0        1   N/A
192.168.100.246 4      64100         1         1        0    0    0       3m             0        1   N/A

Displayed neighbors 2
Total number of neighbors 2
```
* Run the agnhost container on the external host:
```
podman run --network=host -d --entrypoint=/agnhost registry.k8s.io/e2e-test-images/agnhost:2.31 netexec --http-port=9090
```
This is used by the E2E later.

* Run the E2E to verify the setup worked:
```
cd ../e2e/
KUBECONFIG=/root/kubeconfig ginkgo -- -external-host-ip="192.168.100.100"
```