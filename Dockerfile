FROM centos:7

COPY dist/tekton-trigger /bin/tekton-trigger
ENTRYPOINT ["/bin/tekton-trigger"]
