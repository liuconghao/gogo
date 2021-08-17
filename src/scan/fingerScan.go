package scan

import (
	"crypto/md5"
	"encoding/hex"
	"getitle/src/utils"
	"net"
	"strings"
)

func fingerScan(result *utils.Result) {
	//如果是http协议,则判断cms,如果是tcp则匹配规则库.暂时不考虑udp
	if strings.HasPrefix(result.Protocol, "http") {
		getHttpCMS(result)
	} else {
		getTCPFrameWork(result)
	}
	return
}

func getHttpCMS(result *utils.Result) {
	for _, finger := range utils.Httpfingers {
		if VersionLevel < finger.Level {
			return
		}
		httpFingerMatch(result, finger)
		if result.Framework != "" {
			return
		}

	}
	return
}

func httpFingerMatch(result *utils.Result, finger utils.Finger) {

	resp := result.Httpresp
	content := result.Content
	//var cookies map[string]string
	if finger.SendData != "" {
		conn := utils.HttpConn(2)
		resp, err := conn.Get(utils.GetURL(result) + finger.SendData)
		if err != nil {
			return
		}
		content = string(utils.GetBody(resp))
		resp.Body.Close()
	}

	if finger.Regexps.Vuln != nil {
		for _, reg := range utils.Compiled[finger.Name+"_vuln"] {
			res := utils.CompileMatch(reg, content)
			if res == "matched" {
				//println("[*] " + res)
				result.Framework = finger.Name
				result.Vuln = finger.Vuln
				return
			} else if res != "" {
				result.HttpStat = "tcp"
				result.Framework = finger.Name + " " + strings.TrimSpace(res)
				result.Vuln = finger.Vuln
				//result.Title = res
				return
			}
		}
	} else if finger.Regexps.HTML != nil {
		for _, html := range finger.Regexps.HTML {
			if strings.Contains(content, html) {
				result.Framework = finger.Name
				return
			}
		}
	} else if finger.Regexps.Regexp != nil {
		for _, reg := range utils.Compiled[finger.Name] {
			res := utils.CompileMatch(reg, content)
			if res == "matched" {
				//println("[*] " + res)
				result.Framework = finger.Name
				return
			} else if res != "" {
				result.Framework = finger.Name + res
				//result.Title = res
				return
			}
		}
	} else if finger.Regexps.Header != nil {
		for _, header := range finger.Regexps.Header {
			if resp == nil {
				if strings.Contains(content, header) {
					result.Framework = finger.Name
					return
				}
			} else {
				headers := utils.GetHeaderstr(resp)
				if strings.Contains(headers, header) {
					result.Framework = finger.Name
					return
				}
			}
		}
		//} else if finger.Regexps.Cookie != nil {
		//	for _, cookie := range finger.Regexps.Cookie {
		//		if resp == nil {
		//			if strings.Contains(content, cookie) {
		//				result.Framework = finger.Name
		//				return
		//			}
		//		} else if cookies[cookie] != "" {
		//			result.Framework = finger.Name
		//			return
		//		}
		//	}
	} else if finger.Regexps.MD5 != nil {
		for _, md5s := range finger.Regexps.MD5 {
			m := md5.Sum([]byte(content))
			if md5s == hex.EncodeToString(m[:]) {
				result.Framework = finger.Name
				return
			}
		}
	}
}

//第一个返回值为详细的版本信息,第二个返回值为规则名字
func getTCPFrameWork(result *utils.Result) {
	// 优先匹配默认端口,第一遍循环只匹配默认端口
	for _, finger := range utils.Tcpfingers[result.Port] {
		tcpFingerMatch(result, finger)
		if result.Framework != "" {
			return
		}
	}

	// 若默认端口未匹配到结果,则匹配全部
	for port, fingers := range utils.Tcpfingers {
		for _, finger := range fingers {
			if port != result.Port {
				tcpFingerMatch(result, finger)
			}
			if result.Framework != "" {
				return
			}
		}
	}

	return
}

func tcpFingerMatch(result *utils.Result, finger utils.Finger) {
	content := result.Content
	var data []byte
	var err error

	// 某些规则需要主动发送一个数据包探测
	if finger.SendData != "" && VersionLevel >= finger.Level {
		var conn net.Conn
		conn, err = utils.TcpSocketConn(utils.GetTarget(result), 2)
		if err != nil {
			return
		}
		data, err = utils.SocketSend(conn, []byte(finger.SendData), 1024)

		// 如果报错为EOF,则需要重新建立tcp连接
		if err != nil {
			return
			//target := GetTarget(result)
			//// 如果对端已经关闭,则本地socket也关闭,并重新建立连接
			//(*result.TcpCon).Close()
			//*result.TcpCon, err = TcpSocketConn(target, time.Duration(2))
			//if err != nil {
			//	return
			//}
			//data, err = SocketSend(*result.TcpCon, []byte(finger.SendData), 1024)
			//
			//// 重新建立链接后再次报错,则跳过该规则匹配
			//if err != nil {
			//	result.Error = err.Error()
			//	return
			//}
		}
	}
	// 如果主动探测有回包,则正则匹配回包内容, 若主动探测没有返回内容,则直接跳过该规则
	if string(data) != "" {
		content = string(data)
	}

	//遍历指纹正则
	for _, reg := range utils.Compiled[finger.Name] {
		res := utils.CompileMatch(reg, content)
		if res == "matched" {
			//println("[*] " + res)
			result.Framework = finger.Name
			return
		} else if res != "" {
			result.HttpStat = finger.Protocol
			result.Framework = finger.Name + " " + strings.TrimSpace(res)
			//result.Title = res
			return
		}
	}
	// 遍历漏洞正则
	for _, reg := range utils.Compiled[finger.Name+"_vuln"] {
		res := utils.CompileMatch(reg, content)
		if res == "matched" {
			//println("[*] " + res)
			result.Framework = finger.Name
			result.Vuln = finger.Vuln
			return
		} else if res != "" {
			result.HttpStat = "tcp"
			result.Framework = finger.Name + " " + strings.TrimSpace(res)
			result.Vuln = finger.Vuln
			//result.Title = res
			return
		}
	}

	return
}

//func matchHTML()