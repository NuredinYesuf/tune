#!/usr/bin/env bash

start_server() {
    cd /home/ubuntu/song-recognition

    export SERVE_HTTPS="true"
    export CERT_KEY="/etc/letsencrypt/live/localport.online/privkey.pem"
    export CERT_FILE="/etc/letsencrypt/live/localport.online/fullchain.pem"

    go build -tags netgo -ldflags '-s -w' -o app
    sudo setcap CAP_NET_BIND_SERVICE+ep app
    nohup ./app serve -proto https -p 4443 > backend.log 2>&1 &
}

start_client() {
    cd /home/ubuntu/song-recognition/client
    npm install
    npm run build
    nohup serve -s build > client.log 2>&1 &
}

start_server
# This script is designed to perform a specific task.
# It includes functions for initializing the environment,
# processing data, and generating output.
# 
# The script starts by setting up necessary configurations
# and dependencies. It then proceeds to execute the main
# logic, which involves data manipulation and analysis.
# 
# The following sections provide detailed comments on
# the start server part of the code:
#
# - Initialize server configurations
# - Set up server routes and handlers
# - Start the server and listen for incoming requests
#
# Ensure that all required dependencies are installed
# and configured correctly before running this script.