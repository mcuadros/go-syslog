package syslog

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"time"

	. "gopkg.in/check.v1"
)

func getServerConfig() *tls.Config {
	capool := x509.NewCertPool()
	if ok := capool.AppendCertsFromPEM([]byte(caCertPEM)); !ok {
		panic("Cannot add cert")
	}

	cert, err := tls.X509KeyPair([]byte(serverCertPEM), []byte(serverKeyPEM))
	if err != nil {
		panic(err)
	}

	config := tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    capool,
	}

	return &config
}

func getClientConfig() *tls.Config {
	capool := x509.NewCertPool()
	if ok := capool.AppendCertsFromPEM([]byte(caCertPEM)); !ok {
		panic("Cannot add cert")
	}

	cert, err := tls.X509KeyPair([]byte(clientCertPEM), []byte(clientKeyPEM))
	if err != nil {
		panic(err)
	}

	config := tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: false,
		ServerName:         "localhost",
		RootCAs:            capool,
	}
	config.Rand = rand.Reader

	return &config
}

func (s *ServerSuite) TestTLS(c *C) {
	handler := new(HandlerMock)
	server := NewServer()
	server.SetFormat(RFC3164)
	server.SetHandler(handler)
	err := server.ListenTCPTLS("127.0.0.1:0", getServerConfig())
	c.Assert(err, IsNil)
	serverAddr := server.listeners[0].Addr().String()

	err = server.Boot()
	c.Assert(err, IsNil)

	go func(server *Server) {
		time.Sleep(100 * time.Millisecond) // give server time to start
		conn, err := tls.Dial("tcp", serverAddr, getClientConfig())
		c.Assert(err, IsNil)
		defer conn.Close()

		_, err = conn.Write([]byte(exampleSyslog + "\n"))
		c.Assert(err, IsNil)
		time.Sleep(100 * time.Millisecond) // give server time to process request
		err = server.Kill()
		c.Assert(err, IsNil)
	}(server)
	server.Wait()

	c.Check(handler.LastLogParts["hostname"], Equals, "hostname")
	c.Check(handler.LastLogParts["tag"], Equals, "tag")
	c.Check(handler.LastLogParts["content"], Equals, "content")
	c.Check(handler.LastLogParts["tls_peer"], Equals, "client")
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleSyslog)))
	c.Check(handler.LastError, IsNil)
}

const (
	serverKeyPEM = `
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCpjqmTSW5igLI3
R9gzuMUkuwO3ZGC5S8LpoorzUuUwBEhbv2yaIda3XPTCaB6KOn4mKs48m1ViIyHr
WkPRPjsRJ0nCiFzgWOKj6rTrupBEy97LiGNoOTdxeIdRb3d2lqNOkVbBu37ncK61
PMj5qV9F9Lg5sG4jPsDoAUhAcTbB+01HxdCRRa2lgLrnAUWoK3s66HGuKbLdiNSK
fnK1sJAYjUtdxz9DazS4hC7AP8iNoIrEo9wBDSIE8PhPPuGuLeaeYHUwJJ6+MnMW
R6ZXtFL/SG89SQuf1LDf4zlthZ7cqS8rtM3I7iTmDfdJSY+wBvHfU2JF1H+luemw
nTOmwEqHAgMBAAECggEBAJjG1d7DfHW+9lW/I3y/EMuewqN9C3YKYK65abADUkTo
pvYcTlO3B8wiMtv0iwgL2lyzly6e29lYRJjWtWKVSw2Ss/BXhDAVhukhczEv4gxL
Eg2cb82aOG3Cp1LmN+MfqjgB1wUq1xbcvl7JTWE/jnvvHAvHAAY75f9mIF8IY8l2
GV0DyA0OJTlnAtYQQVuxMJ8NrH+JseFaqmk991gH4q95hTfVgBHvqJ6r1L+58pGR
IHS5BLpy5xUfFQOCnPJ3RZhME6zzkmXYRv3sS8+cFJOhlFviJaPLBlvlH9gpkFDC
Is8i5bhxio9xQHOkEQJhhet2N9rc8rzCsA28sI+wAnECgYEA0kK5ClS+f+9VqALZ
1cjbfwJL0PweddDRnCdR1NgnLIIZxVbv6xjpzpTxfhOahKOqDivyACxi3n71eWZC
A6Pl3Rs57mFQjrFr4dBa/rz0n/QhajJln5Nm1yla5kRS3dTPGHK0xcXKeXD4Rb9S
4h/a2thfqLmSdAkozQeR5gGzzNkCgYEAznEyMQXNstlAlqorfRYaZWJc+s+k4J4u
oqYVKjJshrgJErrGu+Fcz7+KMhHxqfM7nuA2CcB5IVcFvUP8aQBf2sw0eSFQVuGq
3+wP+rc4ApIEPy9I6Xi1GFTdwZHIt2exfthczBtDh4XNXoYiIa42jcy3WvLHS34X
tpFp9SFttl8CgYB6og/qxqKVW7JJ29/RoOTknyI5MdNSRAj9WrGPwsKWYwtE3f/w
zwcPRi/TqPtmgU6eFWOAVmMUAliKBepa1S0sWMThFEE3+KNDgZKRIQRMhsc2eU5s
VDyXIbeytgbe+1AOolhtQX9mdU1Y4M4mtQ2gtrKUZifVJcJ2UwP1cui7gQKBgAgt
a7ONa0x+VpShQP+/dGQ3tT8qInnTSj2fHo+BV9MuTw2y4FRo5OhFyg+ZrlzxCZeN
ghZ4zVOIwu1wV/tAzIs6M4noy+nlHoOoMinYQBu59PkbwmOdKG9CTVZxk+XP8bP4
lhRvsAkaP7xSy99Rq0+KoGi13TccU4wjznKrVFE5AoGAN4B31tzm78CYv1BVWVfY
QFc+/Rs65KVmAaQ5Mya2z49jci25xJWBmC+R4IbHFXnfIWFup/m2qdxCcfcFpTBQ
YDPtwEMOI6hwdSjjpVXknb4d7sRDs3HTbUARfR0TGV+KSEHKbcnAjI3yuCrnGgU3
ZEbao7tve6HgmpWi07s4YSI=
-----END PRIVATE KEY-----
`
	// Server Certificate:
	//    Data:
	//        Version: 3 (0x2)
	//        Signature Algorithm: sha256WithRSAEncryption
	//        Issuer: CN=TEST-ONLY-CA
	//        Validity
	//            Not Before: May  3 02:16:01 2021 GMT
	//            Not After : May  1 02:16:01 2031 GMT
	//        Subject: CN=localhost
	//        Subject Public Key Info:
	//            Public Key Algorithm: rsaEncryption
	//                RSA Public-Key: (2048 bit)
	//            X509v3 Subject Alternative Name:
	//                IP Address:127.0.0.1, DNS:localhost
	serverCertPEM = `
-----BEGIN CERTIFICATE-----
MIIDdjCCAl6gAwIBAgIRAJAShFB0l1rYfsDgnx2awO0wDQYJKoZIhvcNAQELBQAw
FzEVMBMGA1UEAwwMVEVTVC1PTkxZLUNBMB4XDTIxMDUwMzAyMTYwMVoXDTMxMDUw
MTAyMTYwMVowFDESMBAGA1UEAwwJbG9jYWxob3N0MIIBIjANBgkqhkiG9w0BAQEF
AAOCAQ8AMIIBCgKCAQEAqY6pk0luYoCyN0fYM7jFJLsDt2RguUvC6aKK81LlMARI
W79smiHWt1z0wmgeijp+JirOPJtVYiMh61pD0T47ESdJwohc4Fjio+q067qQRMve
y4hjaDk3cXiHUW93dpajTpFWwbt+53CutTzI+alfRfS4ObBuIz7A6AFIQHE2wftN
R8XQkUWtpYC65wFFqCt7Ouhxrimy3YjUin5ytbCQGI1LXcc/Q2s0uIQuwD/IjaCK
xKPcAQ0iBPD4Tz7hri3mnmB1MCSevjJzFkemV7RS/0hvPUkLn9Sw3+M5bYWe3Kkv
K7TNyO4k5g33SUmPsAbx31NiRdR/pbnpsJ0zpsBKhwIDAQABo4G/MIG8MAkGA1Ud
EwQCMAAwHQYDVR0OBBYEFOJfY8E/C8vB1kOA5WhlTKNY70QkMFIGA1UdIwRLMEmA
FH7l5M/6ORudJjgvTxfBUZKbXsYyoRukGTAXMRUwEwYDVQQDDAxURVNULU9OTFkt
Q0GCFE72qXaggK6UzUI5jPyGNKt2igEuMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAsG
A1UdDwQEAwIFoDAaBgNVHREEEzARhwR/AAABgglsb2NhbGhvc3QwDQYJKoZIhvcN
AQELBQADggEBAEfiwGwORokJplu0O4xvA98zQWhehALJSsBxC0x3HYhDrMgc9ksG
GUtHOKvKJTnA6SMtpDddfh+LN0YUh2Fwm6IxO1IVoEWa9v8cyF6fZqUrtTCFeD2/
TCIdO4c3A1PqZwh4AxtwWvgKFi3BEXsYJ2lRA1lP/hsHvr+bjt3qTEaoDU7DJEO5
k5h0xA9Q5XkW1dDlWaqQP+/5+f1ch9++AM9PrpOBE3uD7Z+ejZ6mgqhfp2F0MQIK
XKjT3PNC/qxi5ZKBna5dTnRDyTa/6OoyzO5B8kILqf4SFO6s0/Xl7Trmm5ancs9B
yEc731BHdlm3F3Q28U+2fvB0R8bdEYbkonk=
-----END CERTIFICATE-----
`

	clientKeyPEM = `
-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC2r9c65XVP+oAq
P5xdu2DqKRgeMu5BiBqx4EpvuWRdJrV+MxTu64xfudovSECzHNeZUKnrj7H+JKFR
a5g5YaORbEoD33vmWAVG+i49n/JxK481GVun6KHraf3MWxMxCJBmZyLEh8M6CrvG
TY4vK4Mlc4ZRkeEQmz227lEbUUeEUvrfHPyYjIDVMR6QoFtAvLXehDEP6p6XzS6u
mD5pS9gtCKTEq8FTelpiEGloaf4nXKpZwh1yUt8xrPwUamhFojBkKa+P7UrP0IWb
HQGpRBxFwebGA04uquWofvBW4NeG+Pj/XZjdRlYDkVNDExcmAe0DplNBUq4+uHfo
gdKAXcYHAgMBAAECggEAGTO+3GglO+hR4AIwfxnHb+ZFZn0eMzokfJ91hV4tA1DA
vu0rGR6zmc0Y1WcBTfRPpd3j4xRKuMWy33mZYWkf2IL43vnorMk9ySHgWS4Ekyow
MmISK+LC26gelB+IUT5eNVJLEJOuEgbDCsNONyGokPUT9ZLLWrAf3mmYFM2ssQtJ
t9RbRkWhsL/0XKWnu0lz/Xq6SjfmDVJy2FU22RqHAzOnOkZR/5vMWtJru/rJGvfv
bCzdfmDhF2+LXf2mtg8h/Ggatfu4HMu5wN7AkfxcNwJ00ltNXyxEWhAwrE5yn06s
bg8tE7wHB7yFXHhmJw0nrKajUywlxrEDJLVEZFfRqQKBgQDmq7y/hUOzZFWi8sz+
vDGVjV9IrMrOSeCHG+Vo0zpCmMIZnbokhiOOmlzpCvFQaNV6ckHOdFYvuv+yIqne
TyvJAOeDqgCLPrB2ZZAFW9EMFRB39+7qWkQ1ozu0HGSt1+LuxPBCPlvGxSnEAbPs
9l7OKlWUBnDkY0wdzZnvCQd0pQKBgQDKv0CM9rd9GV9mwmFhh/4iwQbJYiC9jaX5
d68XcF7E2Gl5AQOOiHtLsRVfSAWaL2uuhyK1AXzT7Mp81yKTUKbSbuTBZrg4blSD
LCK7x28056YGx4HHvdDMjZ1Qbn9AUJ4L6ApQu9LH/p9d2FtEl3YUEPJJjm1xF728
Od8NIRUUOwKBgEw+a67qP4xmF6A6nON+FO2XwuzkoEw7QwmlgNh7KQCmOVH6PnKg
G9Sg1SD6SvUHEbjdVz8EWRCBwM6Cgp9Gj/RqZhuw72kXGYCo5UfAJ4LU25Kr0r6H
g5AvGibYU7baatn9ImTi87bpqHpvDae/b2q5t3ur/VigMaKQONc3ps05AoGBALL7
WUHX/y25u2WczZjrE+ecXaBkNyD/LflntbNMaOz/W0UOJxSp2aZ9Yq+lhgSSPk5p
T7NY59iyXiMNTKGd/lcgvGMbih+PDp5p1RPOQJcEtKWhdClfoTcjATBjC4U8Zfl+
07RnyvDxD8Ep4ZBQ4VVfjHRw/p5q5f2HXSha/x/HAoGBALSdtyHNPJLsUqou14oQ
lbnRgbuRDOyHxDmgHr5oAVEaTt088jyPQJRAZOZFLAZH5G1pJ5Xkca3FB/TmJwDa
sRzzNIcYhZ5dUZZBap6FOqhzNAkQhYBAHEH7qNg2A4KDCleZLTti+IB+3wDwF+U3
iNVFLPM8+L5NVhYZnzr3OsYa
-----END PRIVATE KEY-----
`

	// Client Certificate:
	//    Data:
	//        Version: 3 (0x2)
	//        Signature Algorithm: sha256WithRSAEncryption
	//        Issuer: CN=TEST-ONLY-CA
	//        Validity
	//            Not Before: May  3 04:01:52 2021 GMT
	//            Not After : May  1 04:01:52 2031 GMT
	//        Subject: CN=client
	//        Subject Public Key Info:
	//            Public Key Algorithm: rsaEncryption
	//                RSA Public-Key: (2048 bit)
	//        X509v3 extensions:
	//            X509v3 Basic Constraints:
	//                CA:FALSE
	//            X509v3 Extended Key Usage:
	//                TLS Web Client Authentication
	//    Signature Algorithm: sha256WithRSAEncryption
	clientCertPEM = `
-----BEGIN CERTIFICATE-----
MIIDVzCCAj+gAwIBAgIRALFKZyAB8AQy/jVcH5n0gxAwDQYJKoZIhvcNAQELBQAw
FzEVMBMGA1UEAwwMVEVTVC1PTkxZLUNBMB4XDTIxMDUwMzA0MDE1MloXDTMxMDUw
MTA0MDE1MlowETEPMA0GA1UEAwwGY2xpZW50MIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAtq/XOuV1T/qAKj+cXbtg6ikYHjLuQYgaseBKb7lkXSa1fjMU
7uuMX7naL0hAsxzXmVCp64+x/iShUWuYOWGjkWxKA9975lgFRvouPZ/ycSuPNRlb
p+ih62n9zFsTMQiQZmcixIfDOgq7xk2OLyuDJXOGUZHhEJs9tu5RG1FHhFL63xz8
mIyA1TEekKBbQLy13oQxD+qel80urpg+aUvYLQikxKvBU3paYhBpaGn+J1yqWcId
clLfMaz8FGpoRaIwZCmvj+1Kz9CFmx0BqUQcRcHmxgNOLqrlqH7wVuDXhvj4/12Y
3UZWA5FTQxMXJgHtA6ZTQVKuPrh36IHSgF3GBwIDAQABo4GjMIGgMAkGA1UdEwQC
MAAwHQYDVR0OBBYEFB0IGxBzTklkqBF5dZkVoMQwfVvIMFIGA1UdIwRLMEmAFH7l
5M/6ORudJjgvTxfBUZKbXsYyoRukGTAXMRUwEwYDVQQDDAxURVNULU9OTFktQ0GC
FE72qXaggK6UzUI5jPyGNKt2igEuMBMGA1UdJQQMMAoGCCsGAQUFBwMCMAsGA1Ud
DwQEAwIHgDANBgkqhkiG9w0BAQsFAAOCAQEANdQWQqi3Ai4YuRkEqHULGjuhRRiF
GIhIOAk478a6+DYvgzwINfP0BkMG657PaLsPC9iBgJMJ66A5t2ILFDD2SZh41jx2
7Hn4svWX1KruUrjNkrD39HMOhEHrv7zbxQRJguQubNepVgdcob1sMwkXVF1+BMl4
SZHinCdoM918NOBAjwLyoAjrFLEDsJ2Hj+bWomzD5ab4LQtdcLVa9id2MAdRTuWv
7iLGmCaM60O7IdZh7DxmR7tU/+sJfuAuQhP1fQ9DNi9J5VY5EAfJ9e4rWfUEXYRC
eInlt0lzAVo2mdju1e22aZmNYCSgQRC3FDEBlSQ5ZLU6X1R5aCpMae9nfQ==
-----END CERTIFICATE-----
`

	// CA Certificate:
	//    Data:
	//        Version: 3 (0x2)
	//        Signature Algorithm: sha256WithRSAEncryption
	//        Issuer: CN = TEST-ONLY-CA
	//        Validity
	//            Not Before: May  3 02:16:01 2021 GMT
	//            Not After : Apr 29 02:16:01 2036 GMT
	//        Subject: CN = TEST-ONLY-CA
	//        Subject Public Key Info:
	//            Public Key Algorithm: rsaEncryption
	//                RSA Public-Key: (2048 bit)
	//        X509v3 extensions:
	//            X509v3 Basic Constraints:
	//                CA:TRUE
	//            X509v3 Key Usage:
	//                Certificate Sign, CRL Sign
	caCertPEM = `-----BEGIN CERTIFICATE-----
MIIDTjCCAjagAwIBAgIUTvapdqCArpTNQjmM/IY0q3aKAS4wDQYJKoZIhvcNAQEL
BQAwFzEVMBMGA1UEAwwMVEVTVC1PTkxZLUNBMB4XDTIxMDUwMzAyMTYwMVoXDTM2
MDQyOTAyMTYwMVowFzEVMBMGA1UEAwwMVEVTVC1PTkxZLUNBMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvxGdCX01h6JdaHebxKopSkMQQhsrESTMtDwT
fwGskkyFnzlMr57OILGHeuNcxlaRKSlL5tZY4BAmDibl+5DXttYn1BRLIqTqfizi
wtl5jpcR2+A6t27qfsszApNaIsqZ8mKgVxDiS6yiql7XKSQ/mjYpOc+Co26XilZx
k69ai3SlHR6aIUKtpRCPiE3Po5x5x/x+NSkfP1CqQ6ldvqx11AtIWR1pItIOBxrz
BWdTigXbr9WMnc7cY6ZI1chUr+HqocisBs68dHdkkZgEc63GdHyF8UIP+7UlmKNq
Qg0+deaiq4kjz6fFoR5qt6LaQu9tH4U16uohQSgKXoGOM9a/eQIDAQABo4GRMIGO
MB0GA1UdDgQWBBR+5eTP+jkbnSY4L08XwVGSm17GMjBSBgNVHSMESzBJgBR+5eTP
+jkbnSY4L08XwVGSm17GMqEbpBkwFzEVMBMGA1UEAwwMVEVTVC1PTkxZLUNBghRO
9ql2oICulM1COYz8hjSrdooBLjAMBgNVHRMEBTADAQH/MAsGA1UdDwQEAwIBBjAN
BgkqhkiG9w0BAQsFAAOCAQEAJPe7PvPj+ET7sCVgCr6c2nyQmlco38Le7j4Q71kR
mtK9iozTiX3XLBlJvLurQx6lhDj+2IFc+JYFMdD7d3tSVBPq4KHCP4jXkR//sjeq
hg0oKIsxpi6wjF8UDXBnWNagqEvn/FW4OR5U0QEz4ei/PgK3MYbb1A9KYPT6389r
IcZRKBhPmI6rFPmxz9eigXH9YqdMrUI0RSvYPSo8smajjHb78e8Z2p0TKB9R7M2v
w3awX8nFu2WLBD7RvLkDYlCH425pjbPlFOHqHXymLyaJkmJfgF4xP5obNKzDFU55
IqZWYDgZFUyDuRjZPAe39JPOP05gXRdVKH28m1IryPwgOA==
-----END CERTIFICATE-----
`
)
