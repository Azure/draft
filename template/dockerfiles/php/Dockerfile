FROM composer:{{ .Config.GetVariableValue "BUILDERVERSION" }} AS build-env
COPY . /app
RUN cd /app && composer install

FROM php:{{ .Config.GetVariableValue "VERSION" }}
ENV PORT {{ .Config.GetVariableValue "PORT" }}
EXPOSE {{ .Config.GetVariableValue "PORT" }}
COPY --from=build-env /app /var/www/html
RUN usermod -u 1000 www-data; \
    a2enmod rewrite; \
    chown -R www-data:www-data /var/www/html
