# development guide

## Prerequisities

- A k8s CLuster(Kubernetes version > v1.18.x < v1.24.x) (could start with [kind](https://github.com/kubernetes-sigs/kind))
- Go (>1.17.x)
- Docker(18.x)

## Step

- fork
- clone
- sync
- develop
- develop in new branch
- push
- pr

### Step1. Fork

1. Visit https://github.com/cuboss/captain
2. Click `Fork` button to create a fork of the project to your Github account

### Step2. Clone fork to local

```shell
# in a working directory (xxx)
$ cd xxx
$ git clone https://github.com/yourgitaccount/captain.git

$ cd captain
# add original repo as upstream remote
$ git remote add upstream https://github.com/cuboss/captain.git

# never push to upstream main or master
$ git remote set-url --push upstream no_push

# confirm your remotes:
$ git remote -v
origin  git@github.com:xxx/captain.git (fetch)
origin  git@github.com:xxx/captain.git (push) 
upstream        git@github.com:cuboss/captain.git (fetch)
upstream        git@github.com:cuboss/captain.git (push)
```

### Step3. keep your branch in sync

```shell
$ git fetch upstream
$ git checkout main
$ git rebase upstream/main
```

### Step4. Add new features or fix issues

### Step5. Development in new branch

### Step6. Push to your fork

### Step7. Create a PR