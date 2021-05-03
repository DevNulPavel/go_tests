package gameserver

import (
    "net"
)

func EqAddressesUDP(addr1, addr2 *net.UDPAddr) bool {
    if addr1.IP.Equal(addr2.IP) && (addr1.Port == addr2.Port) {
        return true
    }
    return false
}
