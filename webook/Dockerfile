FROM ubuntu:22.04
COPY webook app.env /app/
# 这会多一个layer 考虑优化
COPY db/* /app/db/migration
WORKDIR /app
CMD ["./webook"]