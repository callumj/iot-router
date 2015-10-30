cd http

go build .
CERT_FOLDER=../certs2 CLIENT_ID=mqttRouter2 ./http

cd ../