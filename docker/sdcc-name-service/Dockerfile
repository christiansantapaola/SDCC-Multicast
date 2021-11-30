# syntax=docker/dockerfile:1

FROM golang:1.16-alpine

WORKDIR /app

RUN apk add git
ADD https://api.github.com/repos/christiansantapaola/SDCC-Multicast/git/refs/heads/main version.json
RUN git clone -b main https://github.com/christiansantapaola/SDCC-Multicast.git && cd SDCC-Multicast && find cmd -type f -exec go install {} \;
EXPOSE 2080
