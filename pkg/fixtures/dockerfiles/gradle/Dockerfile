FROM gradle:jdk21 as BUILD

COPY --chown=gradle:gradle . /project
RUN gradle -i -s -b /project/build.gradle clean build

FROM eclipse-temurin:21-jre
ENV PORT 80
EXPOSE 80

COPY --from=BUILD /project/build/libs/* /opt/
WORKDIR /opt/
RUN ls -l
CMD ["/bin/bash", "-c", "find -type f -name '*SNAPSHOT.jar' | xargs java -jar"]
