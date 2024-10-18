FROM python:{{ .Config.GetVariableValue "VERSION" }}
ENV PORT {{ .Config.GetVariableValue "PORT" }}
EXPOSE {{ .Config.GetVariableValue "PORT" }}
WORKDIR /usr/src/app

COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

ENTRYPOINT ["python"]
CMD ["{{ .Config.GetVariableValue "ENTRYPOINT" }}"]