#!/bin/bash

# Get existing repo tag for origin/release/*
tag=$(git describe)

# Make sure remote repo is added; may already exist if run before
git remote add gh-public https://github.com/thycotic/dsv-cli.git

# Switch to the public branch
git fetch gh-public
git checkout gh-public/master --force

# Merge with our existing origin/release/* top level commit
git merge --allow-unrelated-histories --squash refs/tags/${tag}

# WARNING: (todo) If a file is deleted in private (origin/master), it may not be removed in #  public (remote/master), so we may orphaned files publicly.
grep -lr '<<<<<<' . | xargs -I_fn -- sh -c 'echo _fn && git checkout --theirs _fn && git add _fn'
git config --global user.name "thycotic-rd"
git config --global user.email "lightvaulttrack@thycotic.com"
git commit -m "Automated from: $SourceVersion"
# Only tag if not RC
if [[ "${tag}" != *"-rc"* ]]
then
    echo "Push Tag ${tag}"
    git tag ${tag} -f
    git push https://thycotic-rd:$githubPat@github.com/thycotic/dsv-cli.git HEAD:master
    git push https://thycotic-rd:$githubPat@github.com/thycotic/dsv-cli.git ${tag}
else
    echo "Tag ${tag} is no allow for push in the remote"
fi