FROM quay.io/opsee/vinz:latest

ENV APPENV ""
ENV HUGS_HOST ""
ENV HUGS_POSTGRES_CONN ""
ENV HUGS_SQS_URL ""
ENV HUGS_AWS_REGION ""
ENV HUGS_OPSEE_HOST ""
ENV HUGS_TEST ""
ENV HUGS_MANDRILL_API_KEY ""
ENV HUGS_VAPE_ENDPOINT ""
ENV HUGS_VAPE_KEYFILE ""
ENV HUGS_LOG_LEVEL ""
ENV HUGS_SLACK_CLIENT_ID ""
ENV HUGS_SLACK_CLIENT_SECRET ""
ENV HUGS_TEST_SLACK_CLIENT_ID ""
ENV HUGS_TEST_SLACK_CLIENT_SECRET ""
ENV HUGS_TEST_SLACK_TOKEN ""
ENV HUGS_OPSEE_HOST ""
ENV HUGS_NOTIFICAPTION_ENDPOINT ""

ENV AWS_ACCESS_KEY_ID ""
ENV AWS_SECRET_ACCESS_KEY ""
ENV AWS_DEFAULT_REGION "us-west-2"
ENV AWS_INSTANCE_ID ""
ENV AWS_SESSION_TOKEN ""

RUN apk add --update bash ca-certificates curl
RUN curl -Lo /opt/bin/migrate https://s3-us-west-2.amazonaws.com/opsee-releases/go/migrate/migrate-linux-amd64 && \
    chmod 755 /opt/bin/migrate

COPY target/linux/amd64/bin/* /
COPY run.sh /run.sh
COPY migrations /migrations
COPY vape.test.key /

EXPOSE 9097
CMD ["/run.sh"]
