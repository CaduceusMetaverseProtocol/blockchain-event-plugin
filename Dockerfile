FROM golang:1.16.4-alpine

ENV service=/go/src/blockchain-event-plugin

COPY . $service/
RUN apk update && apk add gcc && apk add g++ \
&& cd $service/ \
&& go build -o main

FROM alpine

ENV service=/go/src/blockchain-event-plugin
COPY --from=0  $service/main /usr/bin
COPY --from=0  $service/config/log.json  /data/config/

WORKDIR /data

CMD ["main"]
