#!/bin/bash

echo $1 $2

if [ -n "$2" ] || [ -z "$1" ] || ([ "$1" != "client" ] && [ "$1" != "server" ]); then
  echo "Usage: $0 {client|server}"
  exit 1
fi

APP_NAME="proxy-golang"
APP_TYPE=$1
CONFIG_FILE="bin/linux/config.json"
DOCKER_NETWORK_NAME="$APP_NAME-network"
DOCKER_NETWORK_SUBNET="172.18.0.0/16"
DOCKER_IMAGE_NAME="$APP_NAME-image-$APP_TYPE"
DOCKER_CONTAINER_NAME="$APP_NAME-container-$APP_TYPE"
if [ "$APP_TYPE" == "server" ]; then
  ADMIN_PORT=86
  DOCKER_CONTAINER_IP="172.18.0.86"
else
  ADMIN_PORT=91
  DOCKER_CONTAINER_IP="172.18.0.91"
fi

echo "Building executable for $APP_TYPE..."
chmod u+x ./build.sh
./build.sh

echo "Create config file for $APP_TYPE..."
cat > $CONFIG_FILE <<EOL
{
  "appType": "$APP_TYPE",
  "serverHost": "172.18.0.86",
  "serverPort": 9876,
  "serverSecret": "fpllngzieyoh43e8",
  "adminPort": $ADMIN_PORT,
  "adminUser": "admin",
  "adminPass": "pass1word2",
  "tcpPorts": [],
  "udpPorts": []
}
EOL

echo "Deleting existing Docker resources..."
docker stop $DOCKER_CONTAINER_NAME
docker rm $DOCKER_CONTAINER_NAME
docker rmi $DOCKER_IMAGE_NAME -f

echo "Creating Docker network..."
docker network create $DOCKER_NETWORK_NAME --subnet=$DOCKER_NETWORK_SUBNET

cd ..

echo "Building Docker image for $APP_TYPE..."
docker build -f build/Dockerfile -t $DOCKER_IMAGE_NAME --build-arg=ADMIN_PORT=$ADMIN_PORT .

echo "Running Docker container for $APP_TYPE..."
docker run -d --name $DOCKER_CONTAINER_NAME --network $DOCKER_NETWORK_NAME --ip $DOCKER_CONTAINER_IP -p $ADMIN_PORT:$ADMIN_PORT $DOCKER_IMAGE_NAME

echo "Currently running Docker containers:"
docker ps
