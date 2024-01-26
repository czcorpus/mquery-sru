FROM czcorpus/kontext-manatee:2.223.6-jammy

RUN apt-get update && apt-get install wget tar curl git bison libpcre3-dev -y \
    && wget https://go.dev/dl/go1.20.6.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go1.20.6.linux-amd64.tar.gz

WORKDIR /opt
RUN git clone https://github.com/czcorpus/manabuild \
    && cd manabuild \
    && export PATH=$PATH:/usr/local/go/bin \
    && make build && make install

COPY . /opt/mquery-sru
WORKDIR /opt/mquery-sru

RUN git config --global --add safe.directory /opt/mquery-sru \
    && export PATH=$PATH:/usr/local/go/bin:/root/go/bin \
    && go get github.com/mna/pigeon \
    && go install github.com/mna/pigeon \
    && make build