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
$ git clone https://github.com/$yourgitaccount/captain.git

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

- create a branch from main (of your repo)

```shell
git checkout -b somefeature
```

- edit code on `somefeature` branch

- do test and build

```shell
make test
make build
```


### Step5. Development in new branch

- sync with upstream
After the test is completed, it is a good practice to keep your local in sync with upstream to avoid conflicts.

```shell
# Rebase your main branch of your local repo.
git checkout main
git rebase upstream/main  # if conflicted, solve it

# Then make your development branch in sync with master branch
git checkout new_feature
git rebase -i main
```

- commit local changes

```shell
git add <file>
git commit -s -m "some description of your changes"
```

### Step6. Push to your fork

```
git push -f origin somefeature 
```

### Step7. Create a PR

- Visit your fork at https://github.com/$yourgitaccount/captain
- Click the Compare & Pull Request button next to your myfeature branch.
- Check out the pull request process for more details and advice.