FROM alpine:3.15.5

ADD bin/captain-server /opt/captain/bin/captain-server

WORKDIR /opt/captain

ENTRYPOINT [ "/opt/captain/bin/captain-server" ]