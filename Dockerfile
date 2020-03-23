FROM gradle:6.2.2-jdk11 AS build

WORKDIR /

COPY . .

RUN gradle --no-daemon fatJar

FROM openjdk:11-jre-slim AS run

WORKDIR /

COPY --from=build /build/libs/Netsoc-Discord-Bot-fat.jar .

CMD exec java -jar $JAVA_OPTIONS /Netsoc-Discord-Bot-fat.jar