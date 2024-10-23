FROM gradle:jdk21 as BUILD

COPY --chown=gradle:gradle . /project
COPY gradlew gradlew
COPY gradle/wrapper gradle/wrapper
RUN chmod +x gradle/wrapper
RUN chmod +x gradlew
RUN ./gradlew -i -s -b /project/build.gradle clean build

FROM eclipse-temurin:21-jre
ENV PORT 80
EXPOSE 80

COPY --from=BUILD /project/build/libs/* /opt/
WORKDIR /opt/
RUN ls -l
CMD ["/bin/bash", "-c", "find -type f -name '*SNAPSHOT.jar' | xargs java -jar"]