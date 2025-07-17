FROM golang:1.24

ENV CGO_ENABLED=0 
RUN export GO111MODULE=on

RUN mkdir /.cache
RUN chmod 777 /.cache

RUN mkdir -p /go
# /cache/download
RUN chmod 777 /go
# /cache/download

ENV GOCACHE=/.cache/go-build

RUN addgroup --gid 1000 apiuser 
RUN useradd  --uid 1000 --gid 1000 -m -d /home/apiuser apiuser

USER apiuser
ENV PATH /app/bin:$PATH
WORKDIR /app

RUN go install github.com/cespare/reflex@latest
