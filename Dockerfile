# build image
FROM golang:1.22 as build-stage

WORKDIR /app

COPY go.mod go.sum ./ 
RUN go mod download && go mod verify

COPY cmd cmd
COPY internal internal
RUN CGO_ENABLED=0 go build ./cmd/backend.go

# deploy image
FROM alpine:3.14 as deploy-stage
WORKDIR /app

COPY --from=build-stage /app/backend backend

ENTRYPOINT [ "./backend" ]
# ENTRYPOINT [ "ls", "-laR" ]

