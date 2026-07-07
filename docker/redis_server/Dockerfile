FROM redis:8.6

COPY redis.conf /usr/local/etc/redis/redis.conf

USER redis

CMD [ "redis-server", "/usr/local/etc/redis/redis.conf" ]
