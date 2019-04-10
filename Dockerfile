FROM alpine:3.9

COPY ./wallet /opt/coins/bin/wallet
COPY ./etc/db/schema.sql /opt/coins/etc/db/schema.sql

WORKDIR /opt/coins
ENTRYPOINT ["bin/wallet"]
