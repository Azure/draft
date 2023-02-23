FROM clojure as BUILD
COPY . /usr/src/app
WORKDIR /usr/src/app
RUN lein ring uberjar

FROM eclipse-temurin:{{VERSION}}

RUN apk update && apk upgrade && apk add bash
ENV PORT {{PORT}}
EXPOSE {{PORT}}
COPY --from=BUILD /usr/src/app/target/*.jar /opt/
WORKDIR /opt
CMD ["/bin/bash", "-c", "find -type f -name '*standalone.jar' | xargs java -jar"]
