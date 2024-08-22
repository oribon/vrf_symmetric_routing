password zebra

debug bgp updates
debug bgp neighbor
debug zebra nht
debug bgp nht
debug bfd peer
ip nht resolve-via-default
ipv6 nht resolve-via-default

log file /tmp/frr.log debugging
log timestamp precision 3
route-map RMAP permit 10
set ipv6 next-hop prefer-global

router bgp 64200
  no bgp network import-check
  no bgp ebgp-requires-policy
  no bgp default ipv4-unicast

  neighbor {{ .Worker0.IP }} bfd
  neighbor {{ .Worker1.IP }} bfd

  neighbor {{ .Worker0.IP }} remote-as 64100
  neighbor {{ .Worker1.IP }} remote-as 64100

  address-family ipv4 unicast
    network 192.220.55.55/32
    neighbor {{ .Worker0.IP }} next-hop-self
    neighbor {{ .Worker0.IP }} activate

    neighbor {{ .Worker1.IP }} next-hop-self
    neighbor {{ .Worker1.IP }} activate
  exit-address-family

