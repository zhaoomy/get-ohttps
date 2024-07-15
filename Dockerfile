FROM ubuntu:22.04

RUN mkdir /app
RUN mkdir /app/www

ADD bin/cmd /app/
WORKDIR /app/www


ENV OHTTPS_TOKEN=
ENV USERNAME=user
ENV PASSWORD=pass

CMD ["/app/cmd" ,"-ohttps_token=${OHTTPS_TOKEN}" ,"-user=${USERNAME}", "-pass=${PASSWORD}"]
