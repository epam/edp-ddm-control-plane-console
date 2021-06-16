FROM alpine

RUN apk update && apk add ca-certificates

ENV USER_UID=1001 \
    USER_NAME=admin-console \
    HOME=/home/admin-console

RUN addgroup --gid ${USER_UID} ${USER_NAME} \
    && adduser --disabled-password --uid ${USER_UID} --ingroup ${USER_NAME} --home ${HOME} ${USER_NAME}

FROM scratch
COPY --from=0 /etc/passwd /etc/passwd
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /go/bin
ENV PWD=/go/bin
ENV PATH=/go/bin

COPY control-plane-console .
COPY static static
COPY templates templates
COPY .env .
COPY locale locale

USER ${USER_UID}

ENTRYPOINT ["control-plane-console"]