FROM docs/base:latest
MAINTAINER Sven Dowideit <SvenDowideit@docker.com> (@SvenDowideit)

# to get the git info for this repo
COPY . /src

# Reset the /docs dir so we can replace the theme meta with the new repo's git info
RUN git reset --hard

RUN grep "VERSION =" /src/version/version.go  | sed 's/.*"\(.*\)".*/\1/' > /docs/VERSION
COPY docs/* /docs/sources/machine/
COPY docs/mkdocs.yml /docs/mkdocs-machine.yml

# Then build everything together, ready for mkdocs
RUN /docs/build.sh
