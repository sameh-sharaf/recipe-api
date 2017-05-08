FROM golang:1.7

RUN mkdir -p /home/app

WORKDIR /home/app

COPY . /home/app

RUN make deps && make

ENTRYPOINT ["./bin/api-test"]

EXPOSE 8080
