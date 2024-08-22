apiVersion: nmstate.io/v1
kind: NodeNetworkConfigurationPolicy
metadata:
  name: vrfpolicy-worker0
spec:
  nodeSelector:
    kubernetes.io/hostname: {{ .Worker0.NodeName }}
  maxUnavailable: 1
  desiredState:
    interfaces:
    - name: {{ .Intf }}vrf 
      type: vrf 
      state: up
      vrf:
        port:
        - {{ .Intf }}
        route-table-id: 1001 
    - name: {{ .Intf }} 
      type: ethernet
      state: up
      ipv4:
        address:
        - ip: {{ .Worker0.IP }}
          prefix-length: 24
        dhcp: false
        enabled: true
    routes: 
      config:
      - destination: 0.0.0.0/0
        metric: 150
        next-hop-address: {{ .SecondaryNetGW }}
        next-hop-interface: {{ .Intf }}
        table-id: 1001
    route-rules: 
      config:
      - ip-to: 172.30.0.0/16
        priority: 998
        route-table: 254 
      - ip-to: 10.132.0.0/14
        priority: 998
        route-table: 254
      - ip-to: 169.254.169.0/29
        priority: 998
        route-table: 254
---
apiVersion: nmstate.io/v1
kind: NodeNetworkConfigurationPolicy
metadata:
  name: vrfpolicy-worker1
spec:
  nodeSelector:
    kubernetes.io/hostname: {{ .Worker1.NodeName }}
  maxUnavailable: 1
  desiredState:
    interfaces:
    - name: {{ .Intf }}vrf 
      type: vrf 
      state: up
      vrf:
        port:
        - {{ .Intf }}
        route-table-id: 1001 
    - name: {{ .Intf }} 
      type: ethernet
      state: up
      ipv4:
        address:
        - ip: {{ .Worker1.IP }}
          prefix-length: 24
        dhcp: false
        enabled: true
    routes: 
      config:
      - destination: 0.0.0.0/0
        metric: 150
        next-hop-address: {{ .SecondaryNetGW }}
        next-hop-interface: {{ .Intf }}
        table-id: 1001
    route-rules: 
      config:
      - ip-to: 172.30.0.0/16
        priority: 998
        route-table: 254 
      - ip-to: 10.132.0.0/14
        priority: 998
        route-table: 254
      - ip-to: 169.254.169.0/29
        priority: 998
        route-table: 254
