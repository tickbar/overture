package overture

import (
	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
	"net"
	"strings"
)

func chooseDNSServer(message *dns.Msg) dnsServer {

	log.Debug("Question: " + message.Question[0].String())

	question_name := message.Question[0].Name[:len(message.Question[0].Name)-1]

	if isQuestionInIPv6(message) && Config.RedirectIPv6Record {
		return Config.AlternativeDNSServer
	}

	for _, domain := range custom_domain_list {

		if strings.HasSuffix(question_name, domain) {
			log.Debug("Matched: Custom domain " + question_name + " " + domain)
			return Config.AlternativeDNSServer
		}
	}

	log.Debug("Domain match fail, try to use primary DNS.")

	return Config.PrimaryDNSServer
}

func matchIPNetwork(response_message *dns.Msg, question_message *dns.Msg, ip_net_list []*net.IPNet) {

	for _, answer := range response_message.Answer {
		if answer.Header().Rrtype != dns.TypeA {
			continue
		}
		log.Debug("Try to match response ip address with IP network.")
		if isIPMatchList(net.ParseIP(answer.(*dns.A).A.String()), ip_net_list) {
			break
		}
		log.Debug("IP network match fail, finally use alternative DNS.")
		err := getResponse(response_message, question_message, Config.AlternativeDNSServer)
		if err != nil {
			log.Warn("Get dns response failed: ", err)
		}
		return
	}

	log.Debug("Finally use primary DNS.")
}

func logAnswer(message *dns.Msg) {

	for _, answer := range message.Answer {
		log.Debug("Answer: " + answer.String())
	}
}

func isIPMatchList(ip net.IP, ip_net_list []*net.IPNet) bool {

	for _, ip_net := range ip_net_list {
		if ip_net.Contains(ip) {
			log.Debug("Matched: IP network " + ip.String() + " " + ip_net.String())
			return true
		}
	}

	return false
}

func isQuestionInIPv6(message *dns.Msg) bool {

	return message.Question[0].Qtype == dns.TypeAAAA
}