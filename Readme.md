# Release Notes Generator

Application helps to generate release notes based on commit ranges.

Features:
- list release notes from specified commit(this value could be taken from app-interface)
- checks whether specified pull request is present in production environment


### Usage
`make run` - run application

`make generate` - generates release notes to `release_notes` folder

### List of environment variables
This variables could be used to run generator.

`REPO`: Name of the repository. Default value is insights-rbac.

`FROM_COMMIT`: Starting commit for generating the commit list. If not provided, commit will be taken from app-interface from deploy.yml (specified in `PATH_TO_DEPLOY_YAML_APP_INTERFACE`)

`IS_PR_IN_PRODUCTION`: A flag indicating whether the pull request is in production. Default value is 0 - this feature is not activated. If value is <pr number> - application will calculate whether PR is in production of not. 

`OWNER`: Name of the repository owner. Default value is RedHatInsights.

`COMMIT_LIST_RANGE`: The maximum number of commits to include in the commit list. Default value is 100.

`ACCESS_TOKEN`: GitHub personal access token used for authentication.(I used fine grained) 

`APP_INTERFACE_NAMESPACE`: Namespace of the RBAC in app-interface. Default value is /services/insights/rbac/namespaces/rbac-prod.yml.

`PATH_TO_APP_INTERFACE`: Path to the app-interface repository. Default value is /Users/liborpichler/Projects/app-interface/.

`PATH_TO_DEPLOY_YAML_APP_INTERFACE`: Path to the deploy Clowder YAML file for the RBAC. Default value is data/services/insights/rbac/deploy-clowder.yml.

If you want to use this app for rbac service - most of the variables are set by default.

Important variable to set are `ACCESS_TOKEN` and `PATH_TO_APP_INTERFACE`.


### Examples

### Get current realease notes

`ACCESS_TOKEN=XXXX make run`

```
Current commit in production: 2768ba40f64d49cbc6c3e569f9c41e1220983849
Fetching commits...Done
====
[2023-03-15] open api spec update - unified style and singular form for create a role
 - https://github.com/RedHatInsights/insights-rbac/pull/830
 - QE
====

====
[2023-03-09] Add DATABASE_HOST, DATABASE_PORT, API_PATH_PREFIX to rbac_server container 
 - https://github.com/RedHatInsights/insights-rbac/pull/823
 - QE
====
....

```

### Check whether PR with number 830 is in production environment
```
IS_PR_IN_PRODUCTION=830 ACCESS_TOKEN=XXX make run
```

```
./release_notes_generator
Current commit in production: 2768ba40f64d49cbc6c3e569f9c41e1220983849
Fetching commits...Done
Searching for commit 2d677e3d4ec484c6bdb87101afac1446fd91927d to check its presence in production...

NO, PR is NOT in production.
```