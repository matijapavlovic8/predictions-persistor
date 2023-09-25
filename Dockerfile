FROM ubuntu:latest
LABEL authors="mpavlovic"

ENTRYPOINT ["top", "-b"]