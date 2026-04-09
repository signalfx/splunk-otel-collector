FROM python:3.14 AS builder

RUN python -m venv script-project

COPY requirements.txt requirements.txt
COPY cron.py cron.py

RUN bash -c "source script-project/bin/activate && pip install -r requirements.txt && pyinstaller --one-file cron.py"

FROM python:3.14 AS final

COPY --from=builder --chmod=0755 dist/cron .

CMD /cron