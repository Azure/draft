FROM mcr.microsoft.com/dotnet/sdk:6.0 AS builder
WORKDIR /app

# caches restore result by copying csproj file separately
COPY *.csproj .
RUN dotnet restore

COPY . .
RUN dotnet publish --output /app/ --configuration Release --no-restore
RUN ls -l

# Stage 2
FROM mcr.microsoft.com/dotnet/aspnet:6.0
WORKDIR /app
COPY --from=builder /app .

RUN ls -l
ENV PORT {{PORT}}
EXPOSE {{PORT}}

ENTRYPOINT dotnet $(cat /app/__assemblyname).dll
