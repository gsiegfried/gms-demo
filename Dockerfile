FROM golang:1.8.1
COPY . /go/src/github.com/gsiegfried/gms-demo
WORKDIR /go/src/github.com/gsiegfried/gms-demo
RUN cd cmd/api && go build .
RUN cd cmd/geo && go build .
RUN cd cmd/profile && go build .
RUN cd cmd/rate && go build .
RUN cd cmd/www && go build .
