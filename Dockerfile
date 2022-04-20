FROM alpine:3.14.1

RUN apk update && apk add ca-certificates go git

ENV USER_UID=1001 \
    USER_NAME=admin-console \
    HOME=/home/admin-console

RUN addgroup --gid ${USER_UID} ${USER_NAME} \
    && adduser --disabled-password --uid ${USER_UID} --ingroup ${USER_NAME} --home ${HOME} ${USER_NAME}

#FROM scratch
#COPY --from=0 /etc/passwd /etc/passwd
#COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
#COPY --from=0 /usr/lib/go/lib/time/zoneinfo.zip /usr/lib/go/lib/time/

WORKDIR /go/bin
#ENV PWD=/go/bin
#ENV PATH=/go/bin
ENV ZONEINFO=/usr/lib/go/lib/time/zoneinfo.zip

COPY control-plane-console .
COPY static static
COPY templates templates
COPY default.env .
COPY locale locale
COPY osplm.ini .

USER ${USER_UID}

CMD ["control-plane-console"]