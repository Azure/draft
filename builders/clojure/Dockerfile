FROM clojure as BUILD
COPY . /usr/src/app
WORKDIR /usr/src/app
RUN lein ring uberjar

FROM openjdk:8-jdk-alpine
ENV PORT {{PORT}}
EXPOSE {{PORT}}
COPY --from=BUILD /usr/src/app/target/*.jar /opt/app.jar
WORKDIR /opt
CMD ["java", "-jar", "app.jar"]
