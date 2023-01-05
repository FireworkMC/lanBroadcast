package lanbroadcast

import (
	"net"
)

// getHostAddr tries to guess the host ip address to use
func getHostAddr(iface string) (ip net.IP, err error) {
	var interfaces []net.Interface

	if interfaces, err = getInterfaces(iface); err != nil {
		return
	}

	var addrs []net.Addr

	for _, ifa := range interfaces {
		if ifa.Flags&net.FlagLoopback != 0 || ifa.Flags&net.FlagUp == 0 {
			continue
		}

		if addrs, err = ifa.Addrs(); err != nil {
			continue
		}

		for _, addr := range addrs {
			if ip, _, err := net.ParseCIDR(addr.String()); err == nil {
				if ip4 := ip.To4(); ip4 != nil {
					return ip4, nil
				}
			}

		}
	}
	return nil, ErrHost
}

func getInterfaces(iface string) (interfaces []net.Interface, err error) {
	if iface != "" {
		var ifa *net.Interface
		if ifa, err = net.InterfaceByName(iface); err != nil {
			return
		}
		interfaces = append(interfaces, *ifa)
	} else {
		interfaces, err = net.Interfaces()
	}

	return
}
