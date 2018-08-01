package main

import (
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"
)

type Result struct {
	IP           net.IP
	IsV6         bool
	IPBits       int
	Network      net.IP
	NetMask      net.IPMask
	NetMaskSize  int
	HostMask     net.IPMask
	HostMaskSize int
	Max          net.IP
	IPCount      *big.Int
	Tags         []string
}

func main() {
	if len(os.Args) != 2 {
		exitUsage()
	}
	if err := report(os.Stdout, os.Args[1]); err != nil {
		exitUsage()
	}
}

func exitUsage() {
	fmt.Fprintln(os.Stderr, "specify a CIDR e.g. 10.20.30.40/22")
	os.Exit(1)
}

func report(out io.Writer, cidr string) error {
	p := func(format string, args ...interface{}) { fmt.Fprintf(out, format, args...) }
	nl := func() { out.Write([]byte("\n")) }

	r, err := calc(cidr)
	if err != nil {
		return err
	}

	var ipWidth string
	var ipVer string
	if r.IsV6 {
		ipWidth = strconv.Itoa(39) // ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff
		ipVer = "IPv6"
	} else {
		ipWidth = strconv.Itoa(15) // 255.255.255.255
		ipVer = "IPv4"
	}

	hostMaskOffset := strings.Repeat(" ", r.NetMaskSize+r.NetMaskSize/8)

	nl()
	p("          CIDR:  %s\n", cidr)
	if len(r.Tags) > 0 {
		p("          Type:  %s\n", strings.Join(r.Tags, ", "))
	}
	nl()
	p("       IP bits:  %-"+ipWidth+"s  %s\n", fmt.Sprintf("%d (%s)", r.IPBits, ipVer), maskLine(r.IPBits))
	p("    IP address:  %-"+ipWidth+"s  %s\n", r.IP, bin(r.IP))
	nl()
	p("  Network bits:  %-"+ipWidth+"s  %s\n", fmt.Sprintf("%d (..../%d)", r.NetMaskSize, r.NetMaskSize), maskLine(r.NetMaskSize))
	p("  Network mask:  %-"+ipWidth+"s  %s\n", net.IP(r.NetMask), bin(net.IP(r.NetMask)))
	nl()
	p("     Host bits:  %-"+ipWidth+"s  %s%s\n", fmt.Sprintf("%d (%d - %d)", r.HostMaskSize, r.IPBits, r.NetMaskSize), hostMaskOffset, maskLine(r.HostMaskSize))
	p("     Host mask:  %-"+ipWidth+"s  %s\n", net.IP(r.HostMask), bin(net.IP(r.HostMask)))
	nl()
	p(" Number of IPs:  %s\n", fmt.Sprintf("%d (2 ^ %d)", r.IPCount, r.HostMaskSize))
	p("      First IP:  %-"+ipWidth+"s  %s\n", r.Network, bin(r.Network))
	p("       Last IP:  %-"+ipWidth+"s  %s\n", r.Max, bin(r.Max))
	nl()
	return nil
}

func calc(cidr string) (Result, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return Result{}, err
	}
	if ipv4 := ip.To4(); ipv4 != nil {
		ip = ipv4 // 16 -> 4 byte slice
	}

	netMask := ipnet.Mask
	netMaskSize, netMaskBits := netMask.Size()
	hostMask := maskComplement(ipnet.Mask)
	hostMaskSize := netMaskBits - netMaskSize

	tags := []string{}
	if ip.IsLoopback() {
		tags = append(tags, "loopback")
	}
	if ip.IsMulticast() {
		tags = append(tags, "multicast")
	}
	if ip.IsLinkLocalMulticast() {
		tags = append(tags, "link local multicast")
	}
	if ip.IsInterfaceLocalMulticast() {
		tags = append(tags, "interface local multicast")
	}
	if ip.IsGlobalUnicast() {
		// tags = append(tags, "global unicast")
	}
	if ip.IsLinkLocalUnicast() {
		tags = append(tags, "link local unicast")
	}
	if ip.IsUnspecified() {
		tags = append(tags, "unspecified")
	}

	return Result{
		IP:           ip,
		IsV6:         len(ip) == 16,
		IPBits:       len(ip) * 8,
		NetMask:      netMask,
		NetMaskSize:  netMaskSize,
		HostMask:     hostMask,
		HostMaskSize: hostMaskSize,
		Network:      ipnet.IP,
		Max:          maxIP(ipnet),
		IPCount:      new(big.Int).Lsh(big.NewInt(1), uint(hostMaskSize)),
		Tags:         tags,
	}, nil
}

func maxIP(network *net.IPNet) net.IP {
	mask := network.Mask
	bcst := make(net.IP, len(network.IP))
	copy(bcst, network.IP)
	for i := 0; i < len(mask); i++ {
		ipIdx := len(bcst) - i - 1
		bcst[ipIdx] = network.IP[ipIdx] | ^mask[len(mask)-i-1]
	}
	return bcst
}

func bin(ip net.IP) string {
	return strings.Join(binaryOctets(ip), " ")
}

func binaryOctets(ip net.IP) []string {
	octets := []string{}
	for i := 0; i < len(ip); i++ {
		octets = append(octets, fmt.Sprintf("%08b", ip[i]))
	}
	return octets
}

func maskLine(n int) string {
	switch n {
	case 0:
		return ""
	case 1:
		return "1"
	case 2:
		return "2 "
	case 3:
		return "|3|"
	case 4:
		return "|4 |"
	default:
		return maskLineDynamic(n)
	}
}

func maskLineDynamic(n int) string {
	len := n - (2 * len("|")) - (2 * len(" ")) - len(strconv.Itoa(n)) + ((n - 1) / 8)
	if len < 0 {
		len = 0
	}
	lineL := strings.Repeat("-", len/2)
	lineR := strings.Repeat("-", len/2+len%2)
	return "|" + lineL + " " + strconv.Itoa(n) + " " + lineR + "|"
}

func maskComplement(m net.IPMask) net.IPMask {
	comp := make(net.IPMask, len(m))
	copy(comp, m)
	for i := 0; i < len(comp); i++ {
		comp[i] = ^comp[i]
	}
	return comp
}
