cd http-responder

rm -rf http-responder
go build .
CERT_FOLDER=../samplecerts CLIENT_ID=mqttRouter1 ./http-responder

cd ../