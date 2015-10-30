package mqtt

import "os"
import "io/ioutil"
import "fmt"
import "crypto/tls"
import "crypto/x509"
import MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"

func NewTlsConfig() *tls.Config {
	certsFolder := os.Getenv("CERT_FOLDER")
	if certsFolder == "" {
		certsFolder = "samplecerts"
	}

	// Import trusted certificates from CAfile.pem.
	// Alternatively, manually add CA certificates to
	// default openssl CA bundle.
	certpool := x509.NewCertPool()
	pemCerts, err := ioutil.ReadFile(certsFolder + "/CAfile.pem")
	if err == nil {
		certpool.AppendCertsFromPEM(pemCerts)
	}

	// Import client certificate/key pair
	cert, err := tls.LoadX509KeyPair(certsFolder+"/client-crt.pem", certsFolder+"/client-key.pem")
	if err != nil {
		panic(err)
	}

	// Just to print out the client certificate..
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		panic(err)
	}

	// Create tls.Config with desired tls properties
	return &tls.Config{
		// RootCAs = certs used to verify server cert.
		RootCAs: certpool,
		// ClientAuth = whether to request cert from server.
		// Since the server is set up for SSL, this happens
		// anyways.
		ClientAuth: tls.NoClientCert,
		// ClientCAs = certs used to validate client cert.
		ClientCAs: nil,
		// InsecureSkipVerify = verify that cert contents
		// match server. IP matches what is in cert etc.
		InsecureSkipVerify: true,
		// Certificates = list of certs client sends to server.
		Certificates: []tls.Certificate{cert},
	}
}

var f = func(client *MQTT.Client, msg MQTT.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func CreateClient() *MQTT.Client {
	tlsconfig := NewTlsConfig()

	opts := MQTT.NewClientOptions()
	opts.AddBroker("ssl://AYWJ0GDG8C6DS.iot.us-east-1.amazonaws.com:8883")
	opts.SetClientID(os.Getenv("CLIENT_ID")).SetTLSConfig(tlsconfig)
	opts.SetDefaultPublishHandler(f)

	// Start the connection
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return c
}
