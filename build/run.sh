#!/bin/bash

APP_NAME="proxy-golang"
ADMIN_PORT_SERVER=86
ADMIN_PORT_CLIENT=91
ADMIN_PORT_LOCAL=80
DOCKER_CONTAINER_IP_SERVER="172.18.0.86"
DOCKER_CONTAINER_IP_CLIENT="172.18.0.91"
DOCKER_CONTAINER_IP_LOCAL="172.18.0.80"
DOCKER_NETWORK_NAME="$APP_NAME-network"
DOCKER_NETWORK_SUBNET="172.18.0.0/16"
DOCKER_IMAGE_NAME_SERVER="$APP_NAME-image-server"
DOCKER_IMAGE_NAME_CLIENT="$APP_NAME-image-client"
DOCKER_IMAGE_NAME_LOCAL="$APP_NAME-image-local"
DOCKER_CONTAINER_NAME_SERVER="$APP_NAME-container-server"
DOCKER_CONTAINER_NAME_CLIENT="$APP_NAME-container-client"
DOCKER_CONTAINER_NAME_LOCAL="$APP_NAME-container-local"

echo "Building executable..."
chmod u+x ./build.sh
./build.sh

echo "Creating config file for server"
cat > "../bin/linux/config-server.json" <<EOL
{
  "appType": "server",
  "serverPort": 9876,
  "serverSecret": "fpllngzieyoh43e8",
  "adminPort": $ADMIN_PORT_SERVER,
  "adminUser": "admin",
  "adminPass": "pass1word2",
  "tcpPorts": [9101],
  "udpPorts": [9102]
}
EOL

echo "Creating config file for client"
cat > "../bin/linux/config-client.json" <<EOL
{
  "appType": "client",
  "serverHost": "$DOCKER_CONTAINER_IP_SERVER",
  "serverPort": 9876,
  "serverSecret": "fpllngzieyoh43e8",
  "adminPort": $ADMIN_PORT_CLIENT,
  "adminUser": "admin",
  "adminPass": "pass1word2",
  "tcpPorts": [9101],
  "udpPorts": [9102]
}
EOL

echo "Creating config file for local"
cat > "../bin/linux/lighttpd.config" <<EOL
server.document-root = "/var/www/"
server.port = $ADMIN_PORT_LOCAL
EOL

echo "Deleting existing Docker resources..."
docker stop $DOCKER_CONTAINER_NAME_SERVER
docker stop $DOCKER_CONTAINER_NAME_CLIENT
docker stop $DOCKER_CONTAINER_NAME_LOCAL
docker rm $DOCKER_CONTAINER_NAME_SERVER
docker rm $DOCKER_CONTAINER_NAME_CLIENT
docker rm $DOCKER_CONTAINER_NAME_LOCAL
docker rmi $DOCKER_IMAGE_NAME_SERVER -f
docker rmi $DOCKER_IMAGE_NAME_CLIENT -f
docker rmi $DOCKER_IMAGE_NAME_LOCAL -f

echo "Creating Docker network..."
docker network create $DOCKER_NETWORK_NAME --subnet=$DOCKER_NETWORK_SUBNET

cd ..

echo "Building Docker image for server..."
docker build -f docker/Dockerfile.server -t $DOCKER_IMAGE_NAME_SERVER --build-arg=ADMIN_PORT=$ADMIN_PORT_SERVER .

echo "Building Docker image for client..."
docker build -f docker/Dockerfile.client -t $DOCKER_IMAGE_NAME_CLIENT --build-arg=ADMIN_PORT=$ADMIN_PORT_CLIENT .

echo "Building Docker image for local..."
docker build -f docker/Dockerfile.local -t $DOCKER_IMAGE_NAME_LOCAL --build-arg=ADMIN_PORT=$ADMIN_PORT_LOCAL .

cd ./docker

echo "Running Docker container for server..."
docker run -d --name $DOCKER_CONTAINER_NAME_SERVER --network $DOCKER_NETWORK_NAME --ip $DOCKER_CONTAINER_IP_SERVER -p $ADMIN_PORT_SERVER:$ADMIN_PORT_SERVER $DOCKER_IMAGE_NAME_SERVER

echo "Running Docker container for client..."
docker run -d --name $DOCKER_CONTAINER_NAME_CLIENT --network $DOCKER_NETWORK_NAME --ip $DOCKER_CONTAINER_IP_CLIENT -p $ADMIN_PORT_CLIENT:$ADMIN_PORT_CLIENT $DOCKER_IMAGE_NAME_CLIENT

echo "Running Docker container for local..."
docker run -d --name $DOCKER_CONTAINER_NAME_LOCAL --network $DOCKER_NETWORK_NAME --ip $DOCKER_CONTAINER_IP_LOCAL -p $ADMIN_PORT_LOCAL:$ADMIN_PORT_LOCAL $DOCKER_IMAGE_NAME_LOCAL

echo "Currently running Docker containers:"
docker ps
