FROM golang:1.8-alpine

COPY . /go/src/github.com/flavio/kube-image-bouncer
WORKDIR /go/src/github.com/flavio/kube-image-bouncer
RUN go build


FROM alpine
WORKDIR /app
RUN adduser -h /app -D web
COPY --from=0 /go/src/github.com/flavio/kube-image-bouncer/kube-image-bouncer /app/

## Cannot use the --chown option of COPY because it's not supported by
## Docker Hub's automated builds :/
RUN chown -R web:web *
USER web
ENTRYPOINT ["./kube-image-bouncer"]
EXPOSE 1323
