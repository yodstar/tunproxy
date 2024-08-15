FROM golang:1.14

ADD build/tunproxy  /usr/local/tunproxy/tunproxy
ADD tunproxy.conf  /usr/local/tunproxy/tunproxy.conf

ENTRYPOINT /usr/local/tunproxy/tunproxy

EXPOSE 18081
