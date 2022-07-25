FROM alpine:3.15.5

ADD bin/captain-apiserver /opt/captain/bin/captain-apiserver

WORKDIR /opt/captain

CMD ["/bin/sh"]