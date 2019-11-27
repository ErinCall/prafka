FROM golang:1.13
COPY build/server /bin/server
ENV PORT 8000
EXPOSE 8000
CMD ["/bin/server"]
