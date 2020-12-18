FROM image-registry.openshift-image-registry.svc:5000/mdtu-ddm-edp-cicd/golang:1.13.15-stretch

ENV USER_UID=1001 \
    USER_NAME=admin-console \
    HOME=/home/admin-console

RUN addgroup --gid ${USER_UID} ${USER_NAME} \
    && adduser --disabled-password --uid ${USER_UID} --ingroup ${USER_NAME} --home ${HOME} ${USER_NAME}

WORKDIR /go/bin

COPY entrypoint .
COPY static static
COPY views views
COPY conf conf
COPY db db

USER ${USER_UID}

ENTRYPOINT ["entrypoint"]