# MailboxService
[![CircleCI](https://circleci.com/gh/fatmalabidi/MailboxService.svg?style=svg&circle-token=242a76f632699b789c7bbb7712faa657e46cb0f3)](https://app.circleci.com/pipelines/github/fatmalabidi/MailboxService)
<a href='https://github.com/jpoles1/gopherbadger' target='_blank'>![gopherbadger-tag-do-not-edit](https://img.shields.io/badge/Go%20Coverage-88%25-brightgreen.svg?longCache=true&style=flat)</a>
Mailbox service is responsible for all the in-app messaging and notifications 

## configuration
use `CONFIGOR_ENV` environment variable to specify which config to use. available configs are 
1. `config.dev.yml` for development environment 
2. `confif.test.yml` for the test environment
3. `confif.prod.yml` for the production environment
please refer to [template file](config/config.template.yml) to see the available field for configuration

## build locally
1. download the dependencies
 
    ```shell  
    make init
    ```
   
2. run the tests

      ```shell  
        make test
      ```
                    
3. generate test coverage badge
 
   ```shell  
    gopherbadger -md="README.md"
   ```
                                     
4. run test coverage and generate coverage report in html page
 
   ```shell  
    make coverage
   ```
   
5. build the server executable:
 
     please note that the default configuration is for Production environment.
 - for `test` or `dev` environment, set the variable `CONFIGOR_ENV` to  `test` :
    
    ```shell  
        CONFIGOR_ENV=test make build-server
    ```
        
  - for production environment:
      
    ```shell  
      make build-server  
    ```     
    
    
### build docker container
Before building docker container you need to  **export the following variables**
 and refer to [Makefile](Makefile) for docker related commands

 ```bash
export export GITHUB_USER=[github-username]
# this can be created from your personal account -> Developer settings -> personal access token
# this token should be read for repos and that's it 
export export GITHUB_TOKEN=[github-user-token] 
# AWS Stuff
# you can get them from ~/.aws/credentials
export AWS_ACCESS_KEY_ID=[aws_access_key_id]
export AWS_SECRET_ACCESS_KEY=[aws_secret_access_key]
export AWS_DEFAULT_REGION=eu-west-1
```

