FROM golang:1.14 AS builder
WORKDIR /app
COPY .   /app
RUN go mod tidy && go mod download && go mod vendor
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ddm-admin-console .


FROM alpine:latest

ENV USER_UID=1001 \
    USER_NAME=admin-console \
    HOME=/home/admin-console

RUN addgroup --gid ${USER_UID} ${USER_NAME} \
    && adduser --disabled-password --uid ${USER_UID} --ingroup ${USER_NAME} --home ${HOME} ${USER_NAME}

WORKDIR /go/bin

COPY --from=builder /app/ddm-admin-console .
COPY --from=builder /app/static static
COPY --from=builder /app/views views
COPY --from=builder /app/conf conf
COPY --from=builder /app/db db

USER ${USER_UID}

ENTRYPOINT ["ddm-admin-console"]