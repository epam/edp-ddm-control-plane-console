FROM nexus-docker-registry.apps.cicd2.mdtu-ddm.projects.epam.com/golang:1.13.15-stretch

ENV USER_UID=1001 \
    USER_NAME=admin-console \
    HOME=/home/admin-console

RUN addgroup --gid ${USER_UID} ${USER_NAME} \
    && adduser --disabled-password --uid ${USER_UID} --ingroup ${USER_NAME} --home ${HOME} ${USER_NAME}

WORKDIR /go/bin

COPY app .
COPY static static
COPY views views
COPY conf conf

USER ${USER_UID}

ENTRYPOINT ["app"]