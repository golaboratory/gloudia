FROM node:20-alpine AS clbuild
WORKDIR /
COPY ./manage-apps .
RUN npm install --save --legacy-peer-deps yarn
RUN yarn
RUN yarn build
 
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /src
COPY ["Cloudia/Cloudia.Api/Cloudia.Api.csproj", "Cloudia/Cloudia.Api/"]
COPY ["Cloudia/Cloudia.Core/Cloudia.Core.csproj", "Cloudia/Cloudia.Core/"]
COPY ["Cloudia/Cloudia.Entity/Cloudia.Entity.csproj", "Cloudia/Cloudia.Entity/"]
COPY ["Connecten.Api/Connecten.Api.csproj", "Connecten.Api/"]
COPY ["Connecten.Core/Connecten.Core.csproj", "Connecten.Core/"]
COPY ["Connecten.Entity/Connecten.Entity.csproj", "Connecten.Entity/"]
RUN dotnet restore "Connecten.Api/Connecten.Api.csproj"
COPY . .
WORKDIR "/src/Connecten.Api"
RUN dotnet build "Connecten.Api.csproj" -c Release -o /app/build
 
FROM build AS publish
WORKDIR "/src/Connecten.Api"
RUN dotnet publish "Connecten.Api.csproj" -c Release -o /app/publish
 
FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS final
WORKDIR /app
COPY --from=publish /app/publish .
COPY --from=clbuild /dist /app/wwwroot
ENTRYPOINT ["dotnet", "Connecten.Api.dll"]