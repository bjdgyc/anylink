package arpdis

// Reference: github.com/malfunkt/arpfox
// TODO now only support IPv4

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var defaultSerializeOpts = gopacket.SerializeOptions{
	FixLengths:       true,
	ComputeChecksums: true,
}

// NewARPRequest creates a bew ARP packet of type "request.
func NewARPRequest(src *Addr, dst *Addr) ([]byte, error) {
	return buildPacket(src, dst, layers.ARPRequest)
}

// NewARPReply creates a new ARP packet of type "reply".
func NewARPReply(src *Addr, dst *Addr) ([]byte, error) {
	return buildPacket(src, dst, layers.ARPReply)
}

// buildPacket creates an template ARP packet with the given source and
// destination.
func buildPacket(src *Addr, dst *Addr, opt uint16) ([]byte, error) {
	ether := layers.Ethernet{
		EthernetType: layers.EthernetTypeARP,
		SrcMAC:       src.HardwareAddr,
		DstMAC:       dst.HardwareAddr,
	}
	arp := layers.ARP{
		AddrType: layers.LinkTypeEthernet,
		Protocol: layers.EthernetTypeIPv4,

		HwAddressSize:   6,
		ProtAddressSize: 4,
		Operation:       opt,

		SourceHwAddress:   src.HardwareAddr,
		SourceProtAddress: src.IP.To4(),

		DstHwAddress:   dst.HardwareAddr,
		DstProtAddress: dst.IP.To4(),
	}

	buf := gopacket.NewSerializeBuffer()
	err := gopacket.SerializeLayers(buf, defaultSerializeOpts, &ether, &arp)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
