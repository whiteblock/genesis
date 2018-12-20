interfaces {
    ethernet eth0 {
        address 172.16.1.1/16
        duplex auto
        hw-id b8:ac:6f:97:df:0c
        smp_affinity auto
        speed auto
    }
    ethernet eth1 {
        address 10.254.1.1/24
        duplex auto
        hw-id a0:36:9f:27:ec:a0
        smp_affinity auto
        speed auto
        vif 101 {
            address 10.1.0.1/30
        }
        vif 102 {
            address 10.1.0.5/30
        }
        vif 103 {
            address 10.1.0.9/30
        }
        vif 104 {
            address 10.1.0.13/30
        }
        vif 105 {
            address 10.1.0.17/30
        }
        vif 106 {
            address 10.1.0.21/30
        }
        vif 107 {
            address 10.1.0.25/30
        }
        vif 108 {
            address 10.1.0.29/30
        }
        vif 109 {
            address 10.1.0.33/30
        }
        vif 110 {
            address 10.1.0.37/30
        }
        vif 111 {
            address 10.1.0.41/30
        }
        vif 112 {
            address 10.1.0.45/30
        }
        vif 113 {
            address 10.1.0.49/30
        }
        vif 114 {
            address 10.1.0.53/30
        }
        vif 115 {
            address 10.1.0.57/30
        }
        vif 116 {
            address 10.1.0.61/30
        }
        vif 117 {
            address 10.1.0.65/30
        }
        vif 118 {
            address 10.1.0.69/30
        }
        vif 119 {
            address 10.1.0.73/30
        }
        vif 120 {
            address 10.1.0.77/30
        }
        vif 121 {
            address 10.1.0.81/30
        }
    }
    ethernet eth2 {
        address 10.254.2.1/24
        duplex auto
        hw-id a0:36:9f:27:ec:a2
        smp_affinity auto
        speed auto
        vif 101 {
            address 10.2.0.1/30
        }
        vif 102 {
            address 10.2.0.5/30
        }
        vif 103 {
            address 10.2.0.9/30
        }
        vif 104 {
            address 10.2.0.13/30
        }
        vif 105 {
            address 10.2.0.17/30
        }
        vif 106 {
            address 10.2.0.21/30
        }
        vif 107 {
            address 10.2.0.25/30
        }
        vif 108 {
            address 10.2.0.29/30
        }
        vif 109 {
            address 10.2.0.33/30
        }
        vif 110 {
            address 10.2.0.37/30
        }
        vif 111 {
            address 10.2.0.41/30
        }
    }
    ethernet eth3 {
        address 10.254.3.1/24
        duplex auto
        hw-id 00:15:17:14:92:9e
        smp_affinity auto
        speed auto
        vif 101 {
            address 10.3.0.1/30
        }
        vif 102 {
            address 10.3.0.5/30
        }
        vif 103 {
            address 10.3.0.9/30
        }
        vif 104 {
            address 10.3.0.13/30
        }
    }
    ethernet eth4 {
        address 10.254.4.1/24
        duplex auto
        hw-id 00:15:17:14:92:9f
        smp_affinity auto
        speed auto
        vif 101 {
            address 10.4.0.1/30
        }
        vif 102 {
            address 10.4.0.5/30
        }
        vif 103 {
            address 10.4.0.9/30
        }
        vif 104 {
            address 10.4.0.13/30
        }
        vif 105 {
            address 10.4.0.17/30
        }
        vif 106 {
            address 10.4.0.21/30
        }
        vif 107 {
            address 10.4.0.25/30
        }
        vif 108 {
            address 10.4.0.29/30
        }
        vif 109 {
            address 10.4.0.33/30
        }
        vif 110 {
            address 10.4.0.37/30
        }
        vif 111 {
            address 10.4.0.41/30
        }
    }
    ethernet eth5 {
        address 10.254.5.1/24
        duplex auto
        hw-id 00:e0:4c:03:0c:8b
        smp_affinity auto
        speed auto
        vif 101 {
            address 10.5.0.1/30
        }
        vif 102 {
            address 10.5.0.5/30
        }
    }
    ethernet eth6 {
        address 10.254.6.1/24
        duplex auto
        hw-id 00:0e:c6:cf:77:a7
        smp_affinity auto
        speed auto
        vif 101 {
            address 10.6.0.1/30
        }
        vif 102 {
            address 10.6.0.5/30
        }
        vif 103 {
            address 10.6.0.9/30
        }
        vif 104 {
            address 10.6.0.13/30
        }
    }
}
}
service {
    ssh {
        port 22
    }
}
system {
    config-management {
    }
    gateway-address 172.16.0.1
    host-name vyos
    login {
        user appo {
            authentication {
                encrypted-password $6$il9/fOXu.z5G/uB$MFkRhrnntCIN1MvYBBKNa5WdDwswfldPIvrUT8bD2Cd5hQqqTz2g7mbO/bVBwLfLJmrJA6CuD8SYybmyKfPmC.
                plaintext-password ""
            }
            level admin
        }
        user vyos {
            authentication {
                encrypted-password $1$3h4PFFee$94WZxUR.bdzGKCfIBUXF/1
                plaintext-password ""
            }
            level admin
        }
    }
    time-zone UTC
}


/* Warning: Do not remove the following line. */
/* === vyatta-config-version: "cluster@1:config-management@1:conntrack-sync@1:conntrack@1:cron@1:dhcp-relay@1:dhcp-server@4:firewall@5:ipsec@4:nat@4:qos@1:quagga@2:system@6:vrrp@1:wanloadbalance@3:webgui@1:webproxy@1:zone-policy@1" === */
/* Release version: VyOS 1.1.8 */

