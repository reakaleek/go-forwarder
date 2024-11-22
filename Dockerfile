FROM scratch
COPY go-forwarder /usr/bin/go-forwarder
ENTRYPOINT [ "/usr/bin/go-forwarder" ]
