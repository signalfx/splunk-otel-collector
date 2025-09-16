FROM nginx:1.29.1

RUN rm /etc/nginx/conf.d/default.conf
COPY status.conf /etc/nginx/conf.d/status.conf
COPY htpasswd /etc/nginx/.htpasswd
