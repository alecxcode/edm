FROM golang:1.17

WORKDIR /edm

COPY . .
COPY ./testfiles/edm-system.cfg /root/.edm/edm-system.cfg

RUN go mod download
RUN go build

EXPOSE 8090

CMD ["./edm", "--consolelog"]
