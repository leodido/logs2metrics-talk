FROM docker.io/golang:1.11.0-stretch as bbbuilder
ADD main.go /go/src/github.com/leodido/logs2metrics-talk/
ADD go.mod /go/src/github.com/leodido/logs2metrics-talk/
ADD go.sum /go/src/github.com/leodido/logs2metrics-talk/
WORKDIR /go/src/github.com/leodido/logs2metrics-talk
ENV GO111MODULE on
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /example .

FROM scratch
COPY --from=bbbuilder /example /example
ENTRYPOINT ["/example"]