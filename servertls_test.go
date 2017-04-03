package syslog

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"time"

	. "gopkg.in/check.v1"
)

func getServerConfig() *tls.Config {
	capool := x509.NewCertPool()
	if ok := capool.AppendCertsFromPEM([]byte(ca_s)); !ok {
		panic("Cannot add cert")
	}

	cert, err := tls.X509KeyPair([]byte(cert1_s), []byte(priv1_s))
	if err != nil {
		panic(err)
	}

	config := tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    capool,
	}
	config.Rand = rand.Reader

	return &config
}

func getClientConfig() *tls.Config {
	capool := x509.NewCertPool()
	if ok := capool.AppendCertsFromPEM([]byte(ca_s)); !ok {
		panic("Cannot add cert")
	}

	cert, err := tls.X509KeyPair([]byte(cert1_s), []byte(priv1_s))
	if err != nil {
		panic(err)
	}

	config := tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: false,
		ServerName:         "dummycert1",
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
	server.ListenTCPTLS("0.0.0.0:5143", getServerConfig())

	server.Boot()
	go func(server *Server) {
		time.Sleep(100 * time.Millisecond)
		conn, err := tls.Dial("tcp", "127.0.0.1:5143", getClientConfig())
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		if _, err := io.WriteString(conn, fmt.Sprintf("%s\n", exampleSyslog)); err != nil {
			panic(err)
		}
		time.Sleep(10 * time.Millisecond) // sleep to handle message async
		server.Kill()
	}(server)
	server.Wait()

	c.Check(handler.LastLogParts["hostname"], Equals, "hostname")
	c.Check(handler.LastLogParts["tag"], Equals, "tag")
	c.Check(handler.LastLogParts["content"], Equals, "content")
	c.Check(handler.LastLogParts["tls_peer"], Equals, "dummycert1")
	c.Check(handler.LastMessageLength, Equals, int64(len(exampleSyslog)))
	c.Check(handler.LastError, IsNil)
}

const (
	priv1_s = `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`

	cert1_s = `-----BEGIN CERTIFICATE-----
MIIFYDCCA0igAwIBAgIBATANBgkqhkiG9w0BAQUFADBbMQswCQYDVQQGEwJYWDEN
MAsGA1UECBMETm9uZTENMAsGA1UEBxMETm9uZTENMAsGA1UEChMETm9uZTENMAsG
A1UECxMETm9uZTEQMA4GA1UEAxMHZHVtbXljYTAgFw0xNTA3MTYxNzQzMzBaGA8y
MTE1MDYyMjE3NDMzMFowTzELMAkGA1UEBhMCWFgxDTALBgNVBAgTBE5vbmUxDTAL
BgNVBAoTBE5vbmUxDTALBgNVBAsTBE5vbmUxEzARBgNVBAMTCmR1bW15Y2VydDEw
ggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQDFR9Fxe3X8fp1a0ex6rvTH
5XKsiP7iD7OMJrRSt0RRHxBBq84NMNQG5CZ4BJTjQCcI7bgKfHTLsq3pket/cr3E
jXtomP2q8T72pCl+D5gwcU4UW4AAA0oNyXeuaJ/DCj6FYHoNBGoP8I0AIAUJ7Phb
g7T6wROX8ttSQzDkjsiYqWLWaUs8x+sh317M6/J9YdkRAdHINgEKRixkMtbaYzJI
4nLWvWXhrMiGyFRysNP6AFkeLKLzWDYFLqfrQ7TEEfCIdIa9OkFPBcum8njgd4bi
smNHP1MwPZz5esh5lZewzvECoyAXi4raVQNIcz7VEGBDIlMwOh2ylIlA7PUVzzdL
N7U3Kr2Gmtmj39Tkn4wvvkK6Z32QOvU0XaEGT7j24Mm1pTdzPF7EY3WnbiGZOYmm
d3p+8nWE5lp1ffgNqW4Xs1cZcQ6h0wHqCAao01wd/CzVDle7KTqyTBe2BYctZ2XK
AZgnap+P2XQX+BUOSeMG5swh2uklqSQy8JdwaWrD6lohdBkwFCq7CWux6wP79gYS
ZUe9EQYmyddrxceJON7XEoE/Dkuyf95jd/vzs0LLlq02UY8a09jkgvaiI0y0cI4f
1+58FOzBHRyHRlShHXT02MPDYRTJ7Rg5DwHS4nKu9r9ozSX8Mf6J5cIA3izGfJg3
fJ+jcnyeHger+fbYryBk7wIDAQABozkwNzAJBgNVHRMEAjAAMAsGA1UdDwQEAwIF
4DAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwDQYJKoZIhvcNAQEFBQAD
ggIBAFCfYO/WpS/2rTp+VUoGvVJmpNfM2mXmo2mPb0KiEQ2tl6aIbdc7XAR3DIvf
CMi919iW3wcMoe2XAMmZFXnGOYcb1x5mVqiHJdkJZonIDlhEJmEiVXL1FNrRNMia
8sFkZPH3opE3nDpSEc9sXgWKgamKqU94OGVGey8Cdg4VeXwWad0Z9Jfh1QV7n+Hy
3/0KUf1qVQbJlPYg1KGxL1F8proNItVuMzv5ZFpGB5HXmEWiwKeY2RL7dWAAEyuT
tKRHWgnaQxlPPjAyCjBKBjSGqHeYrikOelXJDeJ7A5q9zpgdx+Xj+hlYUqhj3rew
l/62mG0o8xkLDqXZfQi5/O0NbER8mpIqUA3T3RzBrl6bWHQ8pnNtDMdglBFxlzEG
Uqy2VBWZkekczWss4j7hAnuUvw3jc9KTs7kQPla2kTpnxdecdntgs80bHbu18AV/
DB3srRMTeJU301/G4QiqVqG/APRNZRZVsh6FMNIyL18hEI4FoZX0muEB8LnIZ+bx
+Uw6Z5awI6Nx9KEMjN8dW79Ml4aycUVVC46XQhTGC4dfzLOlYHzPitorlrR2oO2E
A1GVZjhGR80m5da8YyghdQ+HMsu5yMSnDeGOFzrqIN/R3JKry7ahwEpC0hwtnlK3
og3xXKOxdcM+zZ4L8yX9imkYpdEPJYqjygETSEvfC2OgU3FQ
-----END CERTIFICATE-----
`
	ca_s = `-----BEGIN CERTIFICATE-----
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
-----END CERTIFICATE-----
`
)
