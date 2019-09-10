#credit goes to - https://github.com/confluentinc/confluent-kafka-go/issues/156#issuecomment-376311118
FROM golang:alpine

RUN apk add --update --no-cache alpine-sdk bash ca-certificates \
      libressl \
      tar \
      git openssh openssl yajl-dev zlib-dev cyrus-sasl-dev openssl-dev build-base coreutils
WORKDIR /root
RUN git clone --branch v1.1.0 https://github.com/edenhill/librdkafka.git
WORKDIR /root/librdkafka
RUN /root/librdkafka/configure
RUN make
RUN make install
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
RUN go get -d -v gopkg.in/confluentinc/confluent-kafka-go.v1/kafka
WORKDIR /go/src/app
COPY *.go ./
RUN go get -d
CMD ["bash"]