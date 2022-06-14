package dbdata

import (
	"errors"
	"net"
	"strings"
	"time"
)

func GetPolicy(Username string) *Policy {
	policyData := &Policy{}
	err := One("Username", Username, policyData)
	if err != nil {
		return policyData
	}
	return policyData
}

func SetPolicy(p *Policy) error {
	var err error
	if p.Username == "" {
		return errors.New("用户名错误")
	}

	// 包含路由
	routeInclude := []ValData{}
	for _, v := range p.RouteInclude {
		if v.Val != "" {
			if v.Val == All {
				routeInclude = append(routeInclude, v)
				continue
			}

			ipMask, _, err := parseIpNet(v.Val)
			if err != nil {
				return errors.New("RouteInclude 错误" + err.Error())
			}

			v.IpMask = ipMask
			routeInclude = append(routeInclude, v)
		}
	}
	p.RouteInclude = routeInclude
	// 排除路由
	routeExclude := []ValData{}
	for _, v := range p.RouteExclude {
		if v.Val != "" {
			ipMask, _, err := parseIpNet(v.Val)
			if err != nil {
				return errors.New("RouteExclude 错误" + err.Error())
			}
			v.IpMask = ipMask
			routeExclude = append(routeExclude, v)
		}
	}
	p.RouteExclude = routeExclude

	// DNS 判断
	clientDns := []ValData{}
	for _, v := range p.ClientDns {
		if v.Val != "" {
			ip := net.ParseIP(v.Val)
			if ip.String() != v.Val {
				return errors.New("DNS IP 错误")
			}
			clientDns = append(clientDns, v)
		}
	}
	if len(routeInclude) == 0 || (len(routeInclude) == 1 && routeInclude[0].Val == "all") {
		if len(clientDns) == 0 {
			return errors.New("默认路由，必须设置一个DNS")
		}
	}
	p.ClientDns = clientDns

	// 域名拆分隧道，不能同时填写
	p.DsIncludeDomains = strings.TrimSpace(p.DsIncludeDomains)
	p.DsExcludeDomains = strings.TrimSpace(p.DsExcludeDomains)
	if p.DsIncludeDomains != "" && p.DsExcludeDomains != "" {
		return errors.New("包含/排除域名不能同时填写")
	}
	// 校验包含域名的格式
	err = CheckDomainNames(p.DsIncludeDomains)
	if err != nil {
		return errors.New("包含域名有误：" + err.Error())
	}
	// 校验排除域名的格式
	err = CheckDomainNames(p.DsExcludeDomains)
	if err != nil {
		return errors.New("排除域名有误：" + err.Error())
	}

	p.UpdatedAt = time.Now()
	if p.Id > 0 {
		err = Set(p)
	} else {
		err = Add(p)
	}

	return err
}
