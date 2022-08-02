FROM mcr.microsoft.com/dotnet/sdk:{{VERSION}} AS builder
WORKDIR /app

# caches restore result by copying csproj file separately
COPY *.csproj .
RUN dotnet restore

COPY . .
RUN dotnet publish --output /app/ --configuration Release --no-restore
RUN sed -n 's:.*<AssemblyName>\(.*\)</AssemblyName>.*:\1:p' *.csproj > __assemblyname
RUN if [ ! -s __assemblyname ]; then filename=$(ls *.csproj); echo ${filename%.*} > __assemblyname; fi

# Stage 2
FROM mcr.microsoft.com/dotnet/aspnet:{{VERSION}}
WORKDIR /app
COPY --from=builder /app .

ENV PORT {{PORT}}
EXPOSE {{PORT}}

ENTRYPOINT dotnet $(cat /app/__assemblyname).dll --urls "http://*:{{PORT}}"
