FROM nginx:latest
RUN set -x \
  && apt-get update \
	&& apt-get install --no-install-recommends --no-install-suggests -y openssl curl ca-certificates \
	&& curl -o /usr/bin/self-signed-tls https://raw.githubusercontent.com/loganstellway/self-signed-ssl/master/self-signed-tls \
	&& apt-get remove --purge --auto-remove -y curl ca-certificates && rm -rf /var/lib/apt/lists/* \
	&& chmod +x /usr/bin/self-signed-tls
COPY helloworld.nginx /etc/nginx/
COPY nginx.sh .
CMD ["./nginx.sh"]
