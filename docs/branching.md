# Branching

The `churl` project uses [OneFlow](https://www.endoflineblog.com/oneflow-a-git-branching-model-and-workflow) with "develop" and "master" branches, with rebase + merge --no-ff.

## Starting a feature branch

```
$ git checkout -b feature/my-feature develop
```

## Completing a feature branch

```
$ git checkout feature/my-feature
$ git rebase -i develop
$ git checkout develop
$ git merge --no-ff feature/my-feature
$ git push origin develop
$ git branch -d feature/my-feature
```

## Starting a release branch

```
$ git checkout -b release/2.3.0 9efc5d
```

## Completing a release branch

```
$ git checkout release/2.3.0
$ git tag 2.3.0
$ git checkout develop
$ git merge release/2.3.0
$ git push --tags origin develop
$ git branch -d release/2.3.0
$ git checkout master
$ git merge --ff-only 2.3.0
```

## Starting a hotfix branch

```
$ git checkout -b hotfix/2.3.1 master
```

## Completing a hotfix branch

```
$ git checkout hotfix/2.3.1
$ git tag 2.3.1
$ git checkout develop
$ git merge hotfix/2.3.1
$ git push --tags origin develop
$ git branch -d hotfix/2.3.1
$ git checkout master
$ git merge --ff-only 2.3.1
```
