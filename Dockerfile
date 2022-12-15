FROM golang:1.18-bullseye AS build

WORKDIR /go/src/project/
COPY . /go/src/project/

RUN CGO_ENABLED=0 go build -trimpath -o /bin/TrafficAlerted

FROM scratch
WORKDIR /
COPY --from=build /bin/TrafficAlerted /bin/TrafficAlerted

EXPOSE 8699:8699
LABEL description="TrafficAlerted"
ENTRYPOINT ["/bin/TrafficAlerted"]