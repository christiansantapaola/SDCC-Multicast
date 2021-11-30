# syntax=docker/dockerfile:1

FROM golang:1.16-alpine

WORKDIR /app

RUN apk add git nano vim
RUN go get -u github.com/christiansantapaola/SDCC-Multicast
# RUN git clone https://github.com/christiansantapaola/SDCC-Multicast && cd SDCC-Multicast && find cmd -type f -exec go install {} \;
ADD https://api.github.com/repos/christiansantapaola/SDCC-Multicast/git/refs/heads/main version.json
RUN git clone -b main https://github.com/christiansantapaola/SDCC-Multicast.git && cd SDCC-Multicast && find cmd -type f -exec go install {} \;

EXPOSE 2079


