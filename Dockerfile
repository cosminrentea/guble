FROM alpine:latest
COPY ./gobbler ./guble-cli/guble-cli /usr/local/bin/
RUN mkdir -p /var/lib/gobbler
VOLUME ["/var/lib/gobbler"]
ENTRYPOINT ["/usr/local/bin/gobbler"]
EXPOSE 8080
