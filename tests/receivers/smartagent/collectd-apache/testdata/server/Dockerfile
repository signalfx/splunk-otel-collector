FROM httpd:2.4-alpine

COPY ./status.conf /usr/local/apache2/conf/extra/status.conf
RUN echo "Include conf/extra/status.conf" >> /usr/local/apache2/conf/httpd.conf
