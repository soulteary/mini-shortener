FROM golang:1.23.0-alpine3.20 AS Builder
WORKDIR /app
ADD * .
RUN go build -ldflags "-w -s" -v .

FROM alpine:3.20
WORKDIR /app
COPY --from=Builder /app/mini-shortener ./
COPY --from=Builder /app/rules          ./
CMD [ "./mini-shortener" ]