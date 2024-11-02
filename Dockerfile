FROM ubuntu:latest

# Install the ca-certificate package
RUN apt-get update && apt-get install -y ca-certificates
# Update the CA certificates in the container
RUN update-ca-certificates

COPY ./audio-server /
RUN chmod +x audio-server

EXPOSE $PORT

ENTRYPOINT ["./audio-server"]
