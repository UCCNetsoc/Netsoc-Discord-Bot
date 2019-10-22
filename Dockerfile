FROM gradle:5.6.2-jdk8 AS build
WORKDIR /
COPY . .
RUN gradle fatJar

FROM openjdk:8-jre-alpine AS run
WORKDIR /
COPY --from=build /build/libs/Netsoc-Discord-Bot-fat.jar .

CMD ["java", "-jar", "/Netsoc-Discord-Bot-fat.jar"]