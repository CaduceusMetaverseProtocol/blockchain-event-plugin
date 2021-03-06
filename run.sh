#!/bin/bash

docker rm -f blockchain-event-plugin
RPC_PORT=18535
MysqlSourceName="root:mysql2022@tcp(47.242.7.7:13306)/cmp_chain?parseTime=true&charset=utf8&loc=Local"
SyncRpcAddr="https://mainnet.block.caduceus.foundation"
docker run -itd -e SyncRpcAddr=$SyncRpcAddr -e RPC_PORT=$RPC_PORT -e MysqlSourceName=$MysqlSourceName --restart=unless-stopped -v /etc/localtime:/etc/localtime -v /etc/timezone:/etc/timezone --name blockchain-event-plugin -v $(pwd)/blockchain-event-plugin:/data  --network=host blockchain-event-plugin

docker logs -f blockchain-event-plugin
