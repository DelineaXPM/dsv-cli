FROM mcr.microsoft.com/devcontainers/base:ubuntu
# BUILD ARGS FROM DEVCONTAINER JSON
ARG DEVCONTAINER_USER
ARG GOPATH

USER root
ENV DOCKER_BUILDKIT=1
ENV GOPATH=$GOPATH
ENV MAGEFILE_ENABLE_COLOR=1
ENV TRUNK_LAUNCHER_QUIET=true
ENV PATH="$GOPATH/bin:/home/$DEVCONTAINER_USER/.local/share/aquaproj-aqua/bin:/home/$DEVCONTAINER_USER/go/bin:$PATH"
# ENV PATH="${GOPATH}/bin:${PATH}" --> this uses HOST path , use the $PATH variable instead
USER $DEVCONTAINER_USER
COPY files/.zshrc /home/$DEVCONTAINER_USER/.zshrc
COPY files/first-run-notice.txt /home/$DEVCONTAINER_USER/first-run-notice.txt

RUN mkdir -p /home/$DEVCONTAINER_USER/.minikube \
  && chown $DEVCONTAINER_USER /home/$DEVCONTAINER_USER/.minikube \
  && /bin/bash -c 'set -euo pipefail && curl https://get.trunk.io -fsSL | bash -s -- -y' \
  && /bin/bash -c 'set -euo pipefail && curl -sSfL https://raw.githubusercontent.com/aquaproj/aqua-installer/v1.0.0/aqua-installer | bash'


VOLUME [ "/var/lib/docker" ]
CMD ["sleep", "infinity"]
# ENTRYPOINT ["/usr/local/share/docker-init.sh"]
