FROM czcorpus/kontext-manatee:2.223.6-jammy

RUN apt-get update && apt-get install wget tar curl git bison libpcre3-dev -y \
    && wget https://go.dev/dl/go1.20.6.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go1.20.6.linux-amd64.tar.gz

WORKDIR /opt/mquery-sru
COPY . .
RUN PATH=$PATH:/usr/local/go/bin:/root/go/bin && ./configure && make build