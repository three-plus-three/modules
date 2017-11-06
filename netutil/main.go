package netutil

import (
	"encoding/binary"
	"errors"
	"net"
	"strings"
)

type IPChecker interface {
	Contains(net.IP) bool
}

type ipRange struct {
	start, end uint32
}

func (r *ipRange) String() string {
	var a, b [4]byte
	binary.BigEndian.PutUint32(a[:], r.start)
	binary.BigEndian.PutUint32(b[:], r.end)
	return net.IP(a[:]).String() + "-" +
		net.IP(b[:]).String()
}

func (r *ipRange) Contains(ip net.IP) bool {
	if ip.To4() == nil {
		return false
	}

	v := binary.BigEndian.Uint32(ip.To4())
	return r.start <= v && v <= r.end
}

func IPRange(start, end net.IP) (IPChecker, error) {
	if start.To4() == nil {
		return nil, errors.New("ip range 不支持 IPv6")
	}
	if end.To4() == nil {
		return nil, errors.New("ip range 不支持 IPv6")
	}
	s := binary.BigEndian.Uint32(start.To4())
	e := binary.BigEndian.Uint32(end.To4())
	return &ipRange{start: s, end: e}, nil
}

func IPRangeWith(start, end string) (IPChecker, error) {
	s := net.ParseIP(start)
	if s == nil {
		return nil, errors.New(start + " is invalid address")
	}
	e := net.ParseIP(end)
	if e == nil {
		return nil, errors.New(end + " is invalid address")
	}
	return IPRange(s, e)
}

var (
	_ IPChecker = &net.IPNet{}
	_ IPChecker = &ipRange{}

	ErrInvalidIPRange = errors.New("invalid ip range")
)

func ToCheckers(ipList []string) ([]IPChecker, error) {
	var ingressIPList []IPChecker
	for _, s := range ipList {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if strings.Contains(s, "-") {
			ss := strings.Split(s, "-")
			if len(ss) != 2 {
				return nil, ErrInvalidIPRange
			}
			checker, err := IPRangeWith(ss[0], ss[1])
			if err != nil {
				return nil, ErrInvalidIPRange
			}
			ingressIPList = append(ingressIPList, checker)
			continue
		}

		if strings.Contains(s, "/") {
			_, cidr, err := net.ParseCIDR(s)
			if err != nil {
				return nil, ErrInvalidIPRange
			}
			ingressIPList = append(ingressIPList, cidr)
			continue
		}

		checker, err := IPRangeWith(s, s)
		if err != nil {
			return nil, ErrInvalidIPRange
		}
		ingressIPList = append(ingressIPList, checker)
	}
	return ingressIPList, nil
}
