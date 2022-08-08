# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.241.1/containers/codespaces-linux/.devcontainer/base.Dockerfile

FROM mcr.microsoft.com/vscode/devcontainers/universal:2-focal


USER root
COPY library-scripts/*.sh /tmp/library-scripts/
RUN apt-get update && /bin/bash /tmp/library-scripts/common-debian.sh "${INSTALL_ZSH}" "${USERNAME}" "${USER_UID}" "${USER_GID}" "${UPGRADE_PACKAGES}" "true" "true" \
    && /bin/bash /tmp/library-scripts/docker-in-docker-debian.sh "${ENABLE_NONROOT_DOCKER}" "${USERNAME}" "${USE_MOBY}" "${DOCKER_VERSION}" \
    && /bin/bash /tmp/library-scripts/install-gcm.sh \
    && apt-get autoremove -y && apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/library-scripts/
VOLUME [ "/var/lib/docker" ]
ENTRYPOINT ["/usr/local/share/docker-init.sh"]
CMD ["sleep", "infinity"]

USER codespace
RUN curl -sfL https://direnv.net/install.sh | bash
RUN go install github.com/magefile/mage@latest
ENV GOPATH=/home/${USERNAME}/go
ENV PATH=/home/${USERNAME}/go/bin:$PATH
ENV SHELL /usr/bin/zsh

USER codespace
RUN python3 -m pip install --user --upgrade pip && python3 -m pip install pre-commit --user