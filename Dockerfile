FROM ubuntu:latest

COPY ./audio-server /
RUN chmod +x audio-server

EXPOSE $PORT

ENTRYPOINT ["./audio-server"]
