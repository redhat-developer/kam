# Dockerfile to bootstrap build and test in openshift-ci

FROM registry.ci.openshift.org/openshift/release:golang-1.14

RUN yum -y install dnf httpd-tools sudo

RUN dnf -y install 'dnf-command(config-manager)'

RUN dnf -y config-manager --add-repo https://cli.github.com/packages/rpm/gh-cli.repo

RUN dnf -y install gh

RUN export HOME="/alabama"