FROM ubuntu:latest

# Install the ca-certificate package
RUN apt-get update && apt-get install -y ca-certificates
# Update the CA certificates in the container
RUN update-ca-certificates

COPY ./audio-server /
RUN chmod +x audio-server

HEALTHCHECK --interval=30s --timeout=5s --retries=3 --start-period=60s \
  CMD curl -f http://localhost:$PORT/actuator/health || exit 1

EXPOSE $PORT

ENTRYPOINT ["./audio-server"]
