FROM apache/spark:3.5.5-python3

COPY ./long_running.py /opt/spark/examples/src/main/python/long_running.py
CMD [ "/opt/spark/bin/spark-submit", "/opt/spark/examples/src/main/python/long_running.py" ]

EXPOSE 4040
