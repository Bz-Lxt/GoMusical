FROM node:20-alpine AS build
WORKDIR /src
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
ARG VITE_API_BASE=/api
ENV VITE_API_BASE=$VITE_API_BASE
COPY frontend-user/package.json frontend-user/package-lock.json* /src/frontend-user/
COPY frontend-creator /src/frontend-creator
COPY frontend-admin /src/frontend-admin
WORKDIR /src/frontend-user
RUN npm install --registry=https://registry.npmmirror.com
COPY frontend-user/ /src/frontend-user/
RUN npm run build

FROM nginx:1.27-alpine
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache wget tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo Asia/Shanghai > /etc/timezone
COPY nginx/default.conf /etc/nginx/conf.d/default.conf
COPY --from=build /src/frontend-user/dist /usr/share/nginx/html
EXPOSE 80
HEALTHCHECK --interval=10s --timeout=3s --retries=5 CMD wget -qO- http://127.0.0.1/ >/dev/null || exit 1
