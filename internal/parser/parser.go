package parser

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// ParseTargets converts a list of IP addresses and CIDR ranges into individual IPs
func ParseTargets(targets []string) ([]string, error) {
	var ips []string
	seen := make(map[string]bool)

	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}

		// Check if it's a CIDR range
		if strings.Contains(target, "/") {
			rangeIPs, err := parseCIDR(target)
			if err != nil {
				return nil, fmt.Errorf("invalid CIDR %s: %w", target, err)
			}
			for _, ip := range rangeIPs {
				if !seen[ip] {
					ips = append(ips, ip)
					seen[ip] = true
				}
			}
		} else {
			// Single IP address
			if net.ParseIP(target) == nil {
				return nil, fmt.Errorf("invalid IP address: %s", target)
			}
			if !seen[target] {
				ips = append(ips, target)
				seen[target] = true
			}
		}
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no valid targets found")
	}

	return ips, nil
}

// parseCIDR expands a CIDR range into individual IP addresses
func parseCIDR(cidr string) ([]string, error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIP(ip) {
		ips = append(ips, ip.String())
	}

	// Remove network and broadcast addresses for typical /24 and smaller
	if len(ips) > 2 {
		ips = ips[1 : len(ips)-1]
	}

	return ips, nil
}

// incrementIP increments an IP address by one
func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// ParsePorts converts a port specification string into individual ports
// Supports formats: "80", "80,443", "80-100", "80,443,8000-9000"
func ParsePorts(portSpec string) ([]int, error) {
	var ports []int
	seen := make(map[int]bool)

	parts := strings.Split(portSpec, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check if it's a range
		if strings.Contains(part, "-") {
			rangePorts, err := parsePortRange(part)
			if err != nil {
				return nil, err
			}
			for _, p := range rangePorts {
				if !seen[p] {
					ports = append(ports, p)
					seen[p] = true
				}
			}
		} else {
			// Single port
			port, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", part)
			}
			if err := validatePort(port); err != nil {
				return nil, err
			}
			if !seen[port] {
				ports = append(ports, port)
				seen[port] = true
			}
		}
	}

	if len(ports) == 0 {
		return nil, fmt.Errorf("no valid ports found")
	}

	return ports, nil
}

// parsePortRange parses a port range like "8000-9000"
func parsePortRange(rangeSpec string) ([]int, error) {
	parts := strings.Split(rangeSpec, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid port range: %s", rangeSpec)
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid start port in range %s: %w", rangeSpec, err)
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid end port in range %s: %w", rangeSpec, err)
	}

	if err := validatePort(start); err != nil {
		return nil, fmt.Errorf("start port: %w", err)
	}
	if err := validatePort(end); err != nil {
		return nil, fmt.Errorf("end port: %w", err)
	}

	if start > end {
		return nil, fmt.Errorf("start port %d is greater than end port %d", start, end)
	}

	if end-start > 65535 {
		return nil, fmt.Errorf("port range too large: %d ports", end-start+1)
	}

	var ports []int
	for p := start; p <= end; p++ {
		ports = append(ports, p)
	}

	return ports, nil
}

// validatePort checks if a port number is valid
func validatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port %d out of valid range (1-65535)", port)
	}
	return nil
}