package network

import (
	"fmt"
	"net"
	"strings"
)

type NetworkInfo struct {
	Interface string `json:"interface"`
	IP        string `json:"ip"`
	IsVPN     bool   `json:"is_vpn"`     // Tailscale or other VPN (100.64.0.0/10)
	IsPrivate bool   `json:"is_private"` // 局域网 (192.168, 10., 172.)
	Label     string `json:"label"`
	URL       string `json:"url"`
}

// GetAvailableIPs 获取本机所有有效的非回环 IPv4 地址，并分类标出 Tailscale / 局域网
func GetAvailableIPs(port int) []NetworkInfo {
	var list []NetworkInfo

	// 始终添加 localhost 作为基准
	list = append(list, NetworkInfo{
		Interface: "lo0",
		IP:        "127.0.0.1",
		IsVPN:     false,
		IsPrivate: false,
		Label:     "本机 (Localhost)",
		URL:       fmt.Sprintf("http://127.0.0.1:%d", port),
	})

	ifaces, err := net.Interfaces()
	if err != nil {
		return list
	}

	for _, iface := range ifaces {
		// 忽略 down 或 loopback 的网卡
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip4 := ip.To4()
			if ip4 == nil {
				continue
			}

			ipStr := ip4.String()
			isTailscale := isTailscaleIP(ip4, iface.Name)
			isPriv := isPrivateIP(ip4)

			label := "局域网 (LAN)"
			if isTailscale {
				label = "Tailscale 内网"
			} else if !isPriv {
				label = "公网 / 其他"
			}

			list = append(list, NetworkInfo{
				Interface: iface.Name,
				IP:        ipStr,
				IsVPN:     isTailscale,
				IsPrivate: isPriv,
				Label:     label,
				URL:       fmt.Sprintf("http://%s:%d", ipStr, port),
			})
		}
	}

	return list
}

// Tailscale CGNAT 地址段为 100.64.0.0/10 (100.64.0.0 – 100.127.255.255) 或者接口名为 tailscale/utun
func isTailscaleIP(ip net.IP, ifaceName string) bool {
	if strings.Contains(strings.ToLower(ifaceName), "tailscale") {
		return true
	}
	// 检查 100.64.0.0/10
	if ip[0] == 100 && (ip[1] >= 64 && ip[1] <= 127) {
		return true
	}
	return false
}

func isPrivateIP(ip net.IP) bool {
	// 10.0.0.0/8
	if ip[0] == 10 {
		return true
	}
	// 172.16.0.0/12
	if ip[0] == 172 && (ip[1] >= 16 && ip[1] <= 31) {
		return true
	}
	// 192.168.0.0/16
	if ip[0] == 192 && ip[1] == 168 {
		return true
	}
	return false
}
