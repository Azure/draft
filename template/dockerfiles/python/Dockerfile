FROM python:{{VERSION}}
ENV PORT {{PORT}}
EXPOSE {{PORT}}
WORKDIR /usr/src/app

COPY requirements.tx[t] ./
RUN [ -f requirements.txt ] && pip install --no-cache-dir -r requirements.txt

COPY . .

ENTRYPOINT ["python"]
CMD ["{{ENTRYPOINT}}"]
