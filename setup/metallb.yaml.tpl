apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: the-local-pool
  namespace: openshift-metallb-system
spec:
  addresses:
  - 192.200.10.1/32
---
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: the-cluster-pool
  namespace: openshift-metallb-system
spec:
  addresses:
  - 192.200.10.2/32
---
apiVersion: metallb.io/v1beta2
kind: BGPPeer
metadata:
  name: the-vrf-peer
  namespace: openshift-metallb-system
spec:
  myASN: 64100
  peerASN: 64200
  peerAddress: {{ .ExternalHostIP }}
  vrf: {{ .Intf }}vrf
  nodeSelectors:
    - matchLabels:
        vrf: "true"
---
apiVersion: metallb.io/v1beta1
kind: BGPAdvertisement
metadata:
  name: the-local-adv
  namespace: openshift-metallb-system
spec:
  ipAddressPools:
    - the-local-pool
  peers:
    - the-vrf-peer 
  nodeSelectors:
    - matchLabels:
        egress-service.k8s.ovn.org/the-namespace-the-local-service: ""
---
apiVersion: metallb.io/v1beta1
kind: BGPAdvertisement
metadata:
  name: the-cluster-adv
  namespace: openshift-metallb-system
spec:
  ipAddressPools:
    - the-cluster-pool
  peers:
    - the-vrf-peer 
  nodeSelectors:
    - matchLabels:
        egress-service.k8s.ovn.org/the-namespace-the-cluster-service: "" 