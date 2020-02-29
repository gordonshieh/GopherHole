package dns

import (
	"blocklist"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var recordsA map[string][]layers.DNSResourceRecord = make(map[string][]layers.DNSResourceRecord)
var recordsAAAA map[string][]layers.DNSResourceRecord = make(map[string][]layers.DNSResourceRecord)

const upstreamDNSHost = "1.1.1.1:53"

func find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func toDNSPacket(data []byte) *layers.DNS {
	packet := gopacket.NewPacket(data, layers.LayerTypeDNS, gopacket.Default)
	dnsLayer := packet.Layer(layers.LayerTypeDNS)
	dnsPacket, _ := dnsLayer.(*layers.DNS)
	return dnsPacket
}

// Server for DNS requests
func Server(bl *blocklist.Blocklist) {
	addr := net.UDPAddr{
		Port: 53,
		IP:   net.ParseIP("0.0.0.0"),
	}
	u, _ := net.ListenUDP("udp", &addr)

	for {
		var data []byte
		var clientAddr net.Addr
		{
			tmp := make([]byte, 1024)
			n, addr, _ := u.ReadFrom(tmp)
			data = tmp[:n]
			clientAddr = addr
		}
		dnsPacket := toDNSPacket(data)
		question := dnsPacket.Questions[0]
		requestType := question.Type
		name := string(question.Name)
		block := bl.ShouldBlockHost(name)

		var cache map[string][]layers.DNSResourceRecord
		switch requestType {
		case layers.DNSTypeA:
			cache = recordsA
		case layers.DNSTypeAAAA:
			cache = recordsAAAA
		default:
			continue
		}

		if block {
			dnsPacket.Answers = nil
			dnsPacket.ANCount = 0

			dnsPacket.QR = true
			dnsPacket.OpCode = layers.DNSOpCodeNotify
			dnsPacket.AA = true
			dnsPacket.ResponseCode = layers.DNSResponseCodeNoErr

			buf := gopacket.NewSerializeBuffer()
			_ = dnsPacket.SerializeTo(buf, gopacket.SerializeOptions{})
			u.WriteTo(buf.Bytes(), clientAddr)
			bl.RecordHistory(&blocklist.HistoryEntry{requestType.String(), clientAddr.String(), name, time.Now(), true})
			continue
		}

		answers, exists := cache[name]
		if exists {
			dnsPacket.Answers = answers
			dnsPacket.ANCount = uint16(len(answers))
			dnsPacket.QR = true
			dnsPacket.OpCode = layers.DNSOpCodeNotify
			dnsPacket.AA = true
			dnsPacket.ResponseCode = layers.DNSResponseCodeNoErr

			buf := gopacket.NewSerializeBuffer()
			_ = dnsPacket.SerializeTo(buf, gopacket.SerializeOptions{})
			u.WriteTo(buf.Bytes(), clientAddr)
		} else {
			var dnsResponse []byte
			{
				upstreamConn, _ := net.Dial("udp", upstreamDNSHost)
				defer upstreamConn.Close()
				upstreamConn.Write(data)
				tmp := make([]byte, 1024)
				udpConn, _ := upstreamConn.(*net.UDPConn)
				upstreamConn.SetReadDeadline(time.Now().Add(time.Second * 1))
				n, _, err := udpConn.ReadFrom(tmp)
				if err != nil {
					println(err)
					panic(err)
				}

				dnsResponse = tmp[:n]
			}
			dnsResponsePacket := toDNSPacket(dnsResponse)
			answers := dnsResponsePacket.Answers
			cache[name] = answers
			u.WriteTo(dnsResponse, clientAddr)
		}

	}

}
