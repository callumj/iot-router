cd twilio-responder

go build .
CERT_FOLDER=../samplecerts CLIENT_ID=mqttRouter1 ./twilio-responder

cd ../