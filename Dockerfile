FROM golang:1.16.4-alpine

ENV project=/go/src/project

COPY . $project/
RUN echo -e 'https://mirrors.aliyun.com/alpine/v3.13/main/\nhttps://mirrors.aliyun.com/alpine/v3.13/community/' > /etc/apk/repositories \
    && apk update && apk add gcc && apk add g++ \
&& cd $project/ \
&& go env -w GOPROXY=https://goproxy.cn,direct \
&& go mod tidy \
&& go build -o main

FROM alpine

ENV project=/go/src/project
COPY --from=0 $project/main /usr/bin
WORKDIR /data

CMD ["main"]
