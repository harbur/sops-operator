FROM scratch
ADD sops-operator /sops-operator
ENTRYPOINT ["/sops-operator"]
