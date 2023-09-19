FROM python:{{VERSION}}
ENV PORT {{PORT}}
EXPOSE {{PORT}}
WORKDIR /usr/src/app

COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

ENTRYPOINT ["python"]
CMD ["{{ENTRYPOINT}}"]