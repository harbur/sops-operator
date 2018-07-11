FROM scratch
ADD project-initializer /project-initializer
ENTRYPOINT ["/project-initializer"]
