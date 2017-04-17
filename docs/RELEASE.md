# Docker Machine Release Process

The Docker Machine release process is fairly straightforward and as many steps
have been taken as possible to make it automated, but there is a procedure and
several "checklist items" which should be documented.  This document is intended
to cover the current Docker Machine release process.  It is written for Docker
Machine core maintainers who might find themselves performing a release.

1.  **Version Bump** -- When all commits for the release have been merged into
master and/or the release branch, submit a pull request bumping the `Version`
variable in `version/version.go` to the release version.  Merge said pull
request.  
2. **Compile Binaries** -- Pull down the latest changes to master and
cross-compile the core binary and plugin binaries for each supported OS /
architecture combination.  As you may notice, this can potentially take a _very_
long time so you may want to use a script like [this one for cross-compiling
them on DigitalOcean in
parallel](https://gist.github.com/nathanleclaire/7f62fc5aa3df19a50f4e).
3. **Archive Binaries** -- The binaries are distributed in `.zip` files so you
need to run the `make release-pack` target to generate the distributed
artifacts.
4. **Upload Archives** -- Use a script or sequence of commands such as [this
one](https://gist.github.com/nathanleclaire/a9bc1f8d60070aeda361) to create a
git tag for the released version, a GitHub release for the released version, and
to upload the released binaries.  At the time of writing the `release` target in
the `Makefile` does not work correctly for this step but it should eventually be
split into a separate target and fixed.
5. **Generate Checksums** -- [This
script](https://gist.github.com/nathanleclaire/c506ad3736d33bd42c2f) will spit
out the checksums for the `.zip` files, which you should copy and paste into the
end of the release notes for anyone who wants to verify the checksums of the
downloaded artifacts.
6. **Add Installation Instructions** -- At the top of the release notes, copy and
paste the installation instructions from the previous release, taking care to
update the referenced download URLs to the new version.
7. **Add Release Notes** -- If release notes are already prepared, copy and
paste them into the release notes section after the installation instructions.
If they are not, you can look through the commits since the previous release
using `git log` and summarize the changes in Markdown, preferably by category.
8. **Update the CHANGELOG.md** -- Add the same notes from the previous step to the
`CHANGELOG.md` file in the repository.
9. **Generate and Add Contributor List** -- It is important to thank our
contributors for their good work and persistence.  Usually I generate a list of
authors involved in the release as a Markdown ordered list using a series of
UNIX commands originating from `git` author information, and paste it into the
release notes..  For instance, to print out the list of all unique authors since
the v0.5.0 release:

        $ git log v0.5.0.. --format="%aN" --reverse | sort | uniq | awk '{printf "- %s\n", $0 }'
        - Amir Mohammad
        - Anthony Dahanne
        - David Gageot
        - Jan Broer
        - Jean-Laurent de Morlhon
        - Kazumichi Yamamoto
        - Kunal Kushwaha
        - Mikhail Zholobov
        - Nate McMaster
        - Nathan LeClaire
        - Olivier Gambier
        - Soshi Katsuta
        - aperepel
        - jviide
        - root

10. **Update the Documentation** -- Ensure that the `docs` branch on GitHub (which
the Docker docs team uses to deploy from) is up to date with the changes to be
deployed from the release branch / master.
11. **Verify the Installation** -- Copy and paste the suggested commands in the
installation notes to ensure that they work properly.  Best of all, grab an
(uninvolved) buddy and have them try it.  `docker-machine -v` should give them
the released version once they have run the install commands.
12. (Optional) **Drink a Glass of Wine** -- You've worked hard on this release.
You deserve it.  For wine suggestions, please consult your friendly neighborhood
sommelier.
