FROM maven:3 as BUILD

COPY . /usr/src/app
RUN mvn --batch-mode -f /usr/src/app/pom.xml clean package

FROM eclipse-temurin:21-jre
ENV PORT 80
EXPOSE 80
COPY --from=BUILD /usr/src/app/target /opt/target
WORKDIR /opt/target

CMD ["/bin/bash", "-c", "find -type f -name '*-SNAPSHOT.jar' | xargs java -jar"]
