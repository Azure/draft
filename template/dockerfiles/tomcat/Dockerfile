FROM tomcat:{{ .Config.GetVariableValue "VERSION" }}

ADD sample.war /usr/local/tomcat/webapps/

EXPOSE {{ .Config.GetVariableValue "PORT" }}

CMD [“catalina.sh”, “run”]