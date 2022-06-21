#!/bin/bash

docker rmi blockchain-event-plugin
docker build . -t blockchain-event-plugin

