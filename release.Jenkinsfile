@Library('whiteblock-dev@validate-tag')
import github.Release
import helm.Chart

String imageName = 'genesis'
String registry = 'gcr.io/infra-dev-249211'

String chartDir = 'chart/genesis'

String gitTagCredentialsId = 'github-repo-pac'
String repo = 'genesis'
// see: https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
def semverRegex = ~/^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$/

pipeline {
  agent any
  environment {
    REV_SHORT = sh(script: "git log --pretty=format:'%h' -n 1", , returnStdout: true).trim()
  }
  parameters {
    string(name: "tag_name", description: "(REQUIRED) The name of the tag.")
    string(
      name: "target_commitish",
      defaultValue: "master",
      description: "Specifies the commitish value that determines where the Git tag is created from."
    )
    text(name: "body", defaultValue: '', description: "Description of the release")
  }
  stages {
    stage('validate tag') {
      steps {
        script {
          def release = new github.Release(
              tag_name: params.tag_name,
              body: params.body,
              target_commitish: params.target_commitish,
              repo: repo
          )
          withCredentials([
            usernameColonPassword(credentialsId: gitTagCredentialsId, variable: 'USERPASS')
          ]) {
            validateTag(release, env.USERPASS)
          }
        }
      }
    }
    stage('publish artifacts') {
      steps {
        script {
          source = new container.Image(
            registry: registry,
            name: imageName,
            /*
            NOTE: Ignores target_commitish value and gets the latest build
                  from master branch no matter what
            */
            tag: "master-${env.REV_SHORT}"
          )
          target = new container.Image(
            registry: registry,
            name: imageName,
            tag: "${params.tag_name}"
          )
          tagContainerImage(source, target)

          // just here for convenience when
          // manually installing chart from source
          target = new container.Image(
            registry: registry,
            name: imageName,
            tag: "latest"
          )
          tagContainerImage(source, target)


          chart = new helm.Chart(
            directory: chartDir,
            version: params.tag_name
          )
          publishHelmChart(chart)
        }
      }
    }
    stage('github release') {
      steps {
        script {
          withCredentials([
            usernameColonPassword(credentialsId: gitTagCredentialsId, variable: 'USERPASS')
          ]) {
            String text = createRelease(release, env.USERPASS)
            println text
          }
        }
      }
    }
  }
}
