FROM golang:1.19.0-alpine3.16 AS Builder
WORKDIR /app
ADD * .
RUN go build -ldflags "-w -s" -v .

FROM alpine:3.16
WORKDIR /app
COPY --from=Builder /app/mini-shortener ./
COPY --from=Builder /app/rules          ./
CMD [ "./mini-shortener" ]