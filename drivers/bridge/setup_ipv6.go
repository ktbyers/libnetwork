package bridge

import (
	"fmt"
	"io/ioutil"
	"net"

	"github.com/vishvananda/netlink"
)

var bridgeIPv6 *net.IPNet

const bridgeIPv6Str = "fe80::1/64"

func init() {
	// We allow ourselves to panic in this special case because we indicate a
	// failure to parse a compile-time define constant.
	if ip, netw, err := net.ParseCIDR(bridgeIPv6Str); err == nil {
		bridgeIPv6 = &net.IPNet{IP: ip, Mask: netw.Mask}
	} else {
		panic(fmt.Sprintf("Cannot parse default bridge IPv6 address %q: %v", bridgeIPv6Str, err))
	}
}

func setupBridgeIPv6(config *NetworkConfiguration, i *bridgeInterface) error {
	// Enable IPv6 on the bridge
	procFile := "/proc/sys/net/ipv6/conf/" + config.BridgeName + "/disable_ipv6"
	if err := ioutil.WriteFile(procFile, []byte{'0', '\n'}, 0644); err != nil {
		return fmt.Errorf("Unable to enable IPv6 addresses on bridge: %v", err)
	}

	_, addrsv6, err := i.addresses()
	if err != nil {
		return err
	}

	if len(addrsv6) == 0 {
		if err := netlink.AddrAdd(i.Link, &netlink.Addr{IPNet: bridgeIPv6}); err != nil {
			return &IPv6AddrAddError{ip: bridgeIPv6, err: err}
		}
	}

	// Store bridge network and default gateway
	i.bridgeIPv6 = bridgeIPv6
	i.gatewayIPv6 = i.bridgeIPv6.IP

	return nil
}

func setupGatewayIPv6(config *NetworkConfiguration, i *bridgeInterface) error {
	if config.FixedCIDRv6 == nil {
		return ErrInvalidContainerSubnet
	}
	if !config.FixedCIDRv6.Contains(config.DefaultGatewayIPv6) {
		return ErrInvalidGateway
	}
	if _, err := ipAllocator.RequestIP(config.FixedCIDRv6, config.DefaultGatewayIPv6); err != nil {
		return err
	}

	// Store requested default gateway
	i.gatewayIPv6 = config.DefaultGatewayIPv6

	return nil
}
